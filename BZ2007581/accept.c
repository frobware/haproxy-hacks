/* accept4/close interceptor */

#define _GNU_SOURCE

//#define HAVE_CLANG_TSA

#ifdef HAVE_CLANG_TSA
#include "tsa.h"
#else
#define TSA(x)		     /* No TSA, make TSA attributes no-ops. */
#define TSA_REQUIRES(x)
#define TSA_GUARDED_BY(x) TSA(guarded_by(x))
#define tsa_mutex pthread_mutex_t
#define tsa_mutex_lock pthread_mutex_lock
#define tsa_mutex_unlock pthread_mutex_unlock
#endif

#include <stdio.h>
#include <arpa/inet.h>
#include <dlfcn.h>
#include <limits.h>
#include <pthread.h>
#include <signal.h>
#include <stdarg.h>
#include <stdlib.h>
#include <string.h>
#include <signal.h>
#include <errno.h>
#include <sys/types.h>
#include <sys/stat.h>
#include <unistd.h>
#include <sys/socket.h>
#include <sys/time.h>
#include <sys/types.h>
#include <sys/un.h>
#include <time.h>

#undef NDEBUG /* always want an assert to fire */
#include <assert.h>

#define NELEMENTS(A) ((sizeof(A) / sizeof(A[0])))
#define SOCKPATH_FMT "/tmp/haproxy-%d.connections"

/* interposed functions. */
static int (*libc_accept)(int sockfd, struct sockaddr *addr,
			  socklen_t *addrlen);
static int (*libc_accept4)(int sockfd, struct sockaddr *addr,
			   socklen_t *addrlen, int flags);
static int (*libc_close)(int fd);
static pid_t (*libc_fork)();

struct intercepted_accept {
	struct timespec accept_time;
	int family;
	int fd; /* sanity check */
	char local_address[INET6_ADDRSTRLEN];
	int local_port;
	char peer_address[INET6_ADDRSTRLEN];
	int peer_port;
};

static tsa_mutex fdtab_lock;
static struct intercepted_accept fdtab[65536 * 2] TSA_GUARDED_BY(&fdtab_lock);
static FILE *fp_dbg;

#define LOCK_FDTAB							\
	do {								\
		if (tsa_mutex_lock(&fdtab_lock) != 0) {			\
			PERROR("error: pthread_mutex_lock: %s\n",	\
			       strerror_r(errno,			\
					  _errbuf, sizeof(_errbuf)));	\
		}							\
	} while (0)

#define UNLOCK_FDTAB							\
	do {								\
		if (tsa_mutex_unlock(&fdtab_lock) != 0) {		\
			PERROR("error: pthread_mutex_unlock: %s\n",	\
			       strerror_r(errno,			\
					  _errbuf, sizeof(_errbuf)));	\
		}							\
	} while (0)

#define ACCEPT_INTERPOSER_DEBUG
#ifdef ACCEPT_INTERPOSER_DEBUG
#define DBG(fmt, ...)						\
	do {							\
		if (fp_dbg) fprintf(fp_dbg, "[%d] %s:%d:%s(): "	\
				    fmt,			\
				    getpid(),			\
				    __FILE__,			\
				    __LINE__,			\
				    __func__,			\
				    ##__VA_ARGS__);		\
	} while (0)

#define PERROR(fmt, ...)					\
	do {							\
		char _errbuf[1024];				\
		if (fp_dbg) fprintf(fp_dbg, "[%d] %s:%d:%s(): "	\
				    fmt,			\
				    getpid(),			\
				    __FILE__,			\
				    __LINE__,			\
				    __func__,			\
				    ##__VA_ARGS__);		\
	} while (0)

#else
#define DBG(fmt, ...)
#define PERROR(fmt, ...)
#endif

/* Return a - b */
static void timespec_sub(const struct timespec *a, const struct timespec *b,
			 struct timespec *result) {
	result->tv_sec = a->tv_sec - b->tv_sec;
	result->tv_nsec = a->tv_nsec - b->tv_nsec;
	if (result->tv_nsec < 0) {
		--result->tv_sec;
		result->tv_nsec += 1000000000L;
	}
}

/* Format time as: days hours mins secs. */
static size_t human_time(const struct timespec *ts, char *s, size_t sz) {
	int days = ts->tv_sec / 60 / 60 / 24;
	int hours = ts->tv_sec / 60 / 60 % 24;
	int mins = ts->tv_sec / 60 % 60;
	int secs = ts->tv_sec % 60;

	return snprintf(s, sz, "%d day%s, %2d hour%s, %2d minute%s, %2d sec%s", days,
			days > 0 ? "s" : "s", hours, hours > 0 ? "s" : "s", mins,
			mins > 0 ? "s" : "s", secs, secs > 0 ? "s" : "s");
}

static void delete_debug_socket(int pid)
{
	struct stat sb;
	char socket_path[PATH_MAX];

	snprintf(socket_path, sizeof(socket_path), SOCKPATH_FMT, pid);

	if (stat(socket_path, &sb) == 0) {
		DBG("unlink %s\n", socket_path);
		if (unlink(socket_path) != 0) {
			PERROR("error: unlink: %s\n", strerror_r(errno, _errbuf, sizeof(_errbuf)));
		}
	}
}

static void exit_handler(int exit_code, void *pidptr)
{
	delete_debug_socket(getpid());
	DBG("exit_code %d\n", exit_code);
}

static int write_connection_state(int fd) TSA_REQUIRES(&fdtab_lock)
{
	int write_error = 0;

	for (size_t i = 0; i < NELEMENTS(fdtab); i++) {
		char msg[4096];
		struct timespec elapsed_time;

		if (fdtab[i].fd == -1) {
			continue;
		}

		if (!(fdtab[i].family == AF_INET || fdtab[i].family == AF_INET6)) {
			continue;
		}

		char timebuf[256] = {'\0'};
		struct timespec current_time = {0, 0};

		clock_gettime(CLOCK_MONOTONIC, &current_time);
		timespec_sub(&current_time, &fdtab[i].accept_time, &elapsed_time);
		human_time(&elapsed_time, timebuf, sizeof(timebuf));

		int n = snprintf(
			msg, sizeof(msg), "pid:%d fd:%d %s%s%s:%d %s%s%s:%d %ld secs (%s)\n",
			getpid(), fdtab[i].fd, fdtab[i].family == AF_INET6 ? "[" : "",
			fdtab[i].local_address, fdtab[i].family == AF_INET6 ? "]" : "",
			fdtab[i].local_port, fdtab[i].family == AF_INET6 ? "[" : "",
			fdtab[i].peer_address, fdtab[i].family == AF_INET6 ? "]" : "",
			fdtab[i].peer_port, elapsed_time.tv_sec, timebuf);
		if (n > 0 && write(fd, msg, n) != n) {
			write_error = -1;
			break;
		}
	}

	return write_error;
}

static void *connection_state_handler(void *userarg)
{
	int listen_fd;
	struct sockaddr_un addr;

	memset(&addr, 0, sizeof(addr));
	addr.sun_family = AF_UNIX;
	snprintf(addr.sun_path, sizeof(addr.sun_path)-1, SOCKPATH_FMT, getpid());

	if ((listen_fd = socket(AF_UNIX, SOCK_STREAM, 0)) == -1) {
		PERROR("error: socket: %s\n", strerror_r(errno, _errbuf, sizeof(_errbuf)));
		return NULL;
	}

	unlink(addr.sun_path);

	if (bind(listen_fd, &addr, sizeof(addr)) != 0) {
		PERROR("error: bind(%s): %s\n", addr.sun_path, strerror_r(errno, _errbuf, sizeof(_errbuf)));
		return NULL;
	}

	if (listen(listen_fd, 5) != 0) {
		PERROR("error: listen: %s\n", strerror_r(errno, _errbuf, sizeof(_errbuf)));
		return NULL;
	}

	DBG("listening for debug connections on %s (fd=%d)\n", addr.sun_path, listen_fd);

	while (1) {
		int cl;

		if ((cl = libc_accept4(listen_fd, NULL, NULL, 0)) == -1) {
			PERROR("error: accept (%s): %s\n", addr.sun_path, strerror_r(errno, _errbuf, sizeof(_errbuf)));
			libc_close(cl);
			continue;
		}

		DBG("accepted new debug client fd %d\n", cl);

		tsa_mutex_lock(&fdtab_lock);
		if (write_connection_state(cl) != 0) {
			PERROR("error: write: %s\n", strerror_r(errno, _errbuf, sizeof(_errbuf)));
		}
		tsa_mutex_unlock(&fdtab_lock);
		libc_close(cl);
	}

	DBG("thread exit\n");
	delete_debug_socket(getpid());
	pthread_exit(NULL);
	return NULL;
}

static __attribute__((constructor (102))) void setup(void)
{
	fp_dbg = stderr;

	if ((libc_fork = dlsym(RTLD_NEXT, "fork")) == NULL) {
		PERROR("error: dlsym(fork): %s\n", strerror_r(errno, _errbuf, sizeof(_errbuf)));
		exit(EXIT_FAILURE); /* has to be fatal */
	}

	if ((libc_accept4 = dlsym(RTLD_NEXT, "accept4")) == NULL) {
		PERROR("error: dlsym(accept4): %s\n", strerror_r(errno, _errbuf, sizeof(_errbuf)));
		exit(EXIT_FAILURE); /* has to be fatal */
	}

	if ((libc_accept = dlsym(RTLD_NEXT, "accept")) == NULL) {
		PERROR("error: dlsym(accept): %s\n", strerror_r(errno, _errbuf, sizeof(_errbuf)));
		exit(EXIT_FAILURE); /* has to be fatal */
	}

	if ((libc_close = dlsym(RTLD_NEXT, "close")) == NULL) {
		PERROR("error: dlsym(close): %s\n", strerror_r(errno, _errbuf, sizeof(_errbuf)));
		exit(EXIT_FAILURE); /* has to be fatal */
	}

	LOCK_FDTAB;
	for (size_t i = 0; i < NELEMENTS(fdtab); i++) {
		fdtab[i].fd = -1; /* cleared */
	}
	UNLOCK_FDTAB;

	if (on_exit(exit_handler, NULL) != 0) {
		abort();
	}

	DBG("ACCEPT-INTERPOSER initialised\n");
	fp_dbg = NULL;
	return;
}

/* libc interposer */
int accept(int sockfd, struct sockaddr *sa, socklen_t *salen)
{
	assert(libc_accept != NULL);
	return accept4(sockfd, sa, salen, 0);
}

/* libc interposer */
int accept4(int sockfd, struct sockaddr *sa, socklen_t *salen, int flags)
{
	if (libc_accept4 == NULL) {
		if ((libc_accept4 = dlsym(RTLD_NEXT, "accept4")) == NULL) {
			PERROR("error: dlsym(accept4): %s\n", strerror_r(errno, _errbuf, sizeof(_errbuf)));
			exit(EXIT_FAILURE); /* has to be fatal */
		}
	}

	struct sockaddr local_addr;
	socklen_t local_addrlen = sizeof(local_addr);
	int clientfd;

	assert(libc_accept4 != NULL);
	clientfd = libc_accept4(sockfd, sa, salen, flags);
	if (clientfd == -1) {
		return clientfd;
	}

	DBG("accept4() = %d\n", clientfd);

	if ((sa == NULL || salen == NULL || *salen == 0) ||
	    (clientfd < 0 || clientfd >= NELEMENTS(fdtab))) {
		return clientfd;
	}

	if (!(sa->sa_family == AF_INET || sa->sa_family == AF_INET6)) {
		return clientfd;
	}

	/* Record peer/local address, current time, et al.  */

	LOCK_FDTAB;
	assert(fdtab[clientfd].fd == -1); /* sanity check; explicitly set in close() */
	memset(&fdtab[clientfd], 0, sizeof(fdtab[clientfd]));
	clock_gettime(CLOCK_MONOTONIC, &fdtab[clientfd].accept_time);
	fdtab[clientfd].family = sa->sa_family;
	fdtab[clientfd].fd = clientfd;

	switch (sa->sa_family) {
	case AF_INET:
		memset(&local_addr, 0, sizeof(local_addr));
		if (getsockname(sockfd, &local_addr, &local_addrlen) == 0) {
			fdtab[clientfd].local_port =
				ntohs(((struct sockaddr_in *)&local_addr)->sin_port);
			inet_ntop(AF_INET, &(((struct sockaddr_in *)&local_addr)->sin_addr),
				  fdtab[clientfd].local_address,
				  sizeof(fdtab[clientfd].local_address));
		}
		fdtab[clientfd].peer_port = ntohs(((struct sockaddr_in *)sa)->sin_port);
		inet_ntop(sa->sa_family, &((struct sockaddr_in *)sa)->sin_addr,
			  fdtab[clientfd].peer_address,
			  sizeof(fdtab[clientfd].peer_address));
		break;
	case AF_INET6:
		memset(&local_addr, 0, sizeof(local_addr));
		if (getpeername(sockfd, &local_addr, &local_addrlen) == 0) {
			fdtab[clientfd].local_port =
				ntohs(((struct sockaddr_in6 *)&local_addr)->sin6_port);
			inet_ntop(AF_INET, &(((struct sockaddr_in6 *)&local_addr)->sin6_addr),
				  fdtab[clientfd].local_address,
				  sizeof(fdtab[clientfd].local_address));
		}
		fdtab[clientfd].peer_port = ntohs(((struct sockaddr_in6 *)sa)->sin6_port);
		inet_ntop(sa->sa_family, &((struct sockaddr_in6 *)sa)->sin6_addr,
			  fdtab[clientfd].peer_address,
			  sizeof(fdtab[clientfd].peer_address));
		break;
	}

	UNLOCK_FDTAB;

	return clientfd;
}

/* libc interposer */
int close(int fd)
{
	if (libc_close == NULL) {
		if ((libc_close = dlsym(RTLD_NEXT, "close")) == NULL) {
			PERROR("error: dlsym(close): %s\n", strerror_r(errno, _errbuf, sizeof(_errbuf)));
			exit(EXIT_FAILURE); /* has to be fatal */
		}
	}

	assert(libc_close != NULL);

	if (fd >= 0 && fd <= NELEMENTS(fdtab)) {
		LOCK_FDTAB;
		fdtab[fd].fd = -1; /* clear */
		UNLOCK_FDTAB;
	}

	int rc = libc_close(fd);
	DBG("close(%d) = %d\n", fd, rc);
	return rc;
}

/* libc interposer */
pid_t fork(void)
{
	if (libc_fork == NULL) {
		if ((libc_fork = dlsym(RTLD_NEXT, "fork")) == NULL) {
			PERROR("error: dlsym(fork): %s\n", strerror_r(errno, _errbuf, sizeof(_errbuf)));
			exit(EXIT_FAILURE); /* has to be fatal */
		}
	}

	assert(libc_fork != NULL);
	pid_t pid = libc_fork();

	if (pid == 0) {
		DBG("fork: child %d\n", getpid());
		pthread_t connection_state_tid;
		if (pthread_create(&connection_state_tid, NULL, &connection_state_handler, NULL) != 0) {
			PERROR("error: pthread_create: %s\n", strerror_r(errno, _errbuf, sizeof(_errbuf)));
		}
	} else if (pid > 0) {
		DBG("fork; cleaning up parent %d\n", getpid());
		delete_debug_socket(getpid());
	}

	return pid;
}
