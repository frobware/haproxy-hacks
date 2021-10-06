/* accept4/close interceptor */

#define _GNU_SOURCE
#include <arpa/inet.h>
#include <dlfcn.h>
#include <limits.h>
#include <pthread.h>
#include <signal.h>
#include <stdarg.h>
#include <stdio.h>
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

extern const char * const sys_errlist[];

/* interposed functions. */
static int (*libc_accept)(int sockfd, struct sockaddr *addr,
			  socklen_t *addrlen);
static int (*libc_accept4)(int sockfd, struct sockaddr *addr,
			   socklen_t *addrlen, int flags);
static int (*libc_close)(int fd);
static sighandler_t (*libc_signal)(int signum, sighandler_t handler);
static void (*libc_exit)(int status);
static void (*libc__exit)(int status);
static int (*pthread_pthread_sigmask)(int how, const sigset_t *set, sigset_t *oldset);
static pid_t (*libc_fork)();

static volatile int terminate;

struct intercepted_accept {
	struct timespec accept_time;
	char local_address[INET6_ADDRSTRLEN];
	int local_port;
	int family;
	int fd; /* sanity check */
	char peer_address[INET6_ADDRSTRLEN];
	int peer_port;
};

static pthread_mutex_t fdtab_lock;
static struct intercepted_accept fdtab[65536 * 2];

static volatile int initialised;
static volatile int listen_fd;
static const FILE *fp_dbg;

#if 1
#define LOCK_FDTAB							\
	do {								\
		if (pthread_mutex_lock(&fdtab_lock) != 0) {		\
			DBG("error: pthread_mutex_lock: %s\n", sys_errlist[errno]); \
		}							\
	} while (0)

#define UNLOCK_FDTAB							\
	do {								\
		if (pthread_mutex_unlock(&fdtab_lock) != 0) {		\
			DBG("error: pthread_mutex_unlock: %s\n", sys_errlist[errno]); \
		}							\
	} while (0)
#else

#define LOCK_FDTAB
#define UNLOCK_FDTAB

#endif


#define DBG(fmt, ...)						\
	do {							\
		if (fp_dbg) fprintf(stderr, "[%d] %s:%d:%s(): "	\
				    fmt,			\
				    getpid(),			\
				    __FILE__,			\
				    __LINE__,			\
				    __func__,			\
				    ##__VA_ARGS__);		\
	} while (0)

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
	char socket_path[PATH_MAX];
	snprintf(socket_path, sizeof(socket_path), "/tmp/haproxy-%d.connections", pid);

	struct stat sb;
	if (stat(socket_path, &sb) == 0) {
		DBG("unlink %s\n", socket_path);
		if (unlink(socket_path) != 0) {
			DBG("error: unlink: %s\n", sys_errlist[errno]);
		}
	}
}

static void exit_handler(int exit_code, void *pidptr)
{
	pid_t pid = getpid();

	if (pidptr != NULL) {
		pid = *(pid_t *)pidptr;
	}
	delete_debug_socket(pid);
	DBG("exit_code %d\n", exit_code);
}

#if 0
static char *write_inetaddr_as_str(const struct sockaddr *sa, char *s, size_t sz, int *port)
{
	switch(sa->sa_family) {
	case AF_INET:
		inet_ntop(AF_INET, &(((struct sockaddr_in *)sa)->sin_addr), s, sz);
		*port = ((struct sockaddr_in *)sa)->sin_port;
		break;

	case AF_INET6:
		inet_ntop(AF_INET6, &(((struct sockaddr_in6 *)sa)->sin6_addr), s, sz);
		*port = ((struct sockaddr_in6 *)sa)->sin6_port;
		break;
	default:
		strncpy(s, "Unknown AF", sz);
		return NULL;
	}

	return s;
}
#endif

static int write_connection_state_locked(int fd) {
	char msg[8192];
	struct timespec elapsed_time;

	LOCK_FDTAB;
	for (size_t i = 0; i < NELEMENTS(fdtab); i++) {
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
		if (n > 0) {
			int wrote = write(fd, msg, n);
			if (wrote != n) {
				return -1;
			}
		}
	}
	UNLOCK_FDTAB;

	return 0;
}

static void *connection_state_handler(void *userarg) {
	struct sockaddr_un addr;

	memset(&addr, 0, sizeof(addr));
	addr.sun_family = AF_UNIX;
	snprintf(addr.sun_path, sizeof(addr.sun_path)-1, "/tmp/haproxy-%d.connections", getpid());

	if ((listen_fd = socket(AF_UNIX, SOCK_STREAM, 0)) == -1) {
		DBG("error: socket: %s\n", sys_errlist[errno]);
		return NULL;
	}

	unlink(addr.sun_path);

	if (bind(listen_fd, &addr, sizeof(addr)) != 0) {
		DBG("error: bind(%s): %s\n", addr.sun_path, sys_errlist[errno]);
		assert(0);
		return NULL;
	}

	if (listen(listen_fd, 5) != 0) {
		DBG("error: listen: %s\n", sys_errlist[errno]);
		assert(0);
		return NULL;
	}

	DBG("listening for debug connections on %d (%s)\n", listen_fd, addr.sun_path);
	pthread_setname_np(pthread_self(), "DEBUG");

	while (1) {
		int cl;

		if ((cl = libc_accept4(listen_fd, NULL, NULL, 0)) == -1) {
			DBG("error: accept (%s): %s\n", addr.sun_path, sys_errlist[errno]);
			libc_close(cl);
			continue;
		}
#if 0
		if (!terminate) {
			break;
		}
#endif
		DBG("accepted new debug client fd %d\n", cl);

		if (write_connection_state_locked(cl) != 0) {
			DBG("error: write: %s\n", sys_errlist[errno]);
		}

		libc_close(cl);
	}

	DBG("thread exit\n");
	delete_debug_socket(getpid());
	pthread_exit(NULL);
	return NULL;
}

static void signal_handler(int signum) {
	abort();
	delete_debug_socket(getpid());

	/* restore default handler and redeliver. */
	signal(signum, SIG_DFL);

	if (raise(signum) != 0) {
		exit(0);
	}
}

static __attribute__((constructor)) void setup(void) {
	assert(initialised == 0);
	fp_dbg = stderr;

	if ((pthread_pthread_sigmask = dlsym(RTLD_NEXT, "pthread_sigmask")) == NULL) {
		DBG("error: dlsym(pthread_sigmask): %s\n", sys_errlist[errno]);
		exit(EXIT_FAILURE); /* has to be fatal */
	}

	if ((libc_fork = dlsym(RTLD_NEXT, "fork")) == NULL) {
		DBG("error: dlsym(fork): %s\n", sys_errlist[errno]);
		exit(EXIT_FAILURE); /* has to be fatal */
	}

	if ((libc_accept4 = dlsym(RTLD_NEXT, "accept4")) == NULL) {
		DBG("error: dlsym(accept4): %s\n", sys_errlist[errno]);
		exit(EXIT_FAILURE); /* has to be fatal */
	}

	if ((libc_accept = dlsym(RTLD_NEXT, "accept")) == NULL) {
		DBG("error: dlsym(accept): %s\n", sys_errlist[errno]);
		exit(EXIT_FAILURE); /* has to be fatal */
	}

	if ((libc_close = dlsym(RTLD_NEXT, "close")) == NULL) {
		DBG("error: dlsym(close): %s\n", sys_errlist[errno]);
		exit(EXIT_FAILURE); /* has to be fatal */
	}

	if ((libc_exit = dlsym(RTLD_NEXT, "exit")) == NULL) {
		DBG("error: dlsym(exit): %s\n", sys_errlist[errno]);
		exit(EXIT_FAILURE); /* has to be fatal */
	}

	if ((libc__exit = dlsym(RTLD_NEXT, "_exit")) == NULL) {
		DBG("error: dlsym(_exit): %s\n", sys_errlist[errno]);
		exit(EXIT_FAILURE); /* has to be fatal */
	}

	if ((libc_signal = dlsym(RTLD_NEXT, "signal")) == NULL) {
		DBG("error: dlsym(signal): %s\n", sys_errlist[errno]);
		exit(EXIT_FAILURE); /* has to be fatal */
	}

	LOCK_FDTAB;
	for (size_t i = 0; i < NELEMENTS(fdtab); i++) {
		fdtab[i].fd = -1; /* cleared */
	}
	UNLOCK_FDTAB;

	/* HAProxy doesn't register SIGTERM or SIGINT. It does register:
	 *  1 (SIGHUP)
	 *  3 (SIGQUIT)
	 * 10 (SIGUSR1)
	 * 12 (SIGUSR2)
	 * 13 (SIGPIPE)
	 * 21 (SIGTTIN)
	 * 22 (SIGTTOU)
	 */

	/*
	 * We register SIGTERM and SIGINT so that we delete_debug_socket the
	 * socket path that is created.
	 */
#if 0
	if (signal(SIGTERM, signal_handler) == SIG_ERR) {
		DBG("error: signal(SIGTERM): %s\n", sys_errlist[errno]);
		return;
	}

	if (signal(SIGINT, signal_handler) == SIG_ERR) {
		DBG("error: signal(SIGINT): %s\n", sys_errlist[errno]);
		return;
	}
#endif
	
	pthread_t connection_state_tid;
	if (pthread_create(&connection_state_tid, NULL, &connection_state_handler, NULL) != 0) {
		DBG("error: pthread_create: %s\n", sys_errlist[errno]);
	}

	if (on_exit(exit_handler, NULL) != 0) {
		abort();
	}

	initialised = 1;
	DBG("ACCEPT-INTERPOSER initialised\n");
	return;
}

/* libc interposer */
int accept(int sockfd, struct sockaddr *sa, socklen_t *salen) {
	assert(libc_accept != NULL);
	return accept4(sockfd, sa, salen, 0);
}

/* libc interposer */
int accept4(int sockfd, struct sockaddr *sa, socklen_t *salen, int flags) {
	assert(libc_accept4 != NULL);

	if (listen_fd == 0) {
		pthread_t connection_state_tid;
		if (pthread_create(&connection_state_tid, NULL, &connection_state_handler,
				   NULL) != 0) {
			DBG("error: pthread_create: %s", sys_errlist[errno]);
		}
	}

	int clientfd = libc_accept4(sockfd, sa, salen, flags);

	if (!initialised || clientfd == -1) {
		return clientfd;
	}

	DBG("accept4() = %d\n", clientfd);

	if ((sa == NULL || salen == NULL || *salen == 0) ||
	    (clientfd < 0 || clientfd >= NELEMENTS(fdtab))) {
		return clientfd;
	}

	LOCK_FDTAB;
	assert(fdtab[clientfd].fd == -1); /* sanity check; set in close() */
	memset(&fdtab[clientfd], 0, sizeof(fdtab[clientfd]));

	struct sockaddr local_addr;
	socklen_t local_addrlen = sizeof(local_addr);

	clock_gettime(CLOCK_MONOTONIC, &fdtab[clientfd].accept_time);

	switch (sa->sa_family) {
	case AF_INET:
		fdtab[clientfd].family = sa->sa_family;
		fdtab[clientfd].fd = clientfd;
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
		fdtab[clientfd].family = sa->sa_family;
		fdtab[clientfd].fd = clientfd;
		if (getsockname(sockfd, &local_addr, &local_addrlen) == 0) {
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
int close(int fd) {
	if (libc_close == NULL || fd == -1) {
		return -1;
	}

	assert(libc_close != NULL);

	if (initialised) {
		if (fd >= 0 && fd <= NELEMENTS(fdtab)) {
			LOCK_FDTAB;
			fdtab[fd].fd = -1; /* clear */
			UNLOCK_FDTAB;
		}
	}

	int rc = libc_close(fd);
	DBG("close(%d) = %d\n", fd, rc);
	return rc;
}

/* libc interposer */
pid_t fork(void) {
	assert(libc_fork != NULL);
	pid_t pid = libc_fork();

	if (pid == 0) {
		DBG("new child %d\n", getpid());
		pthread_t connection_state_tid;
		if (pthread_create(&connection_state_tid, NULL, &connection_state_handler, NULL) != 0) {
			DBG("error: pthread_create: %s\n", sys_errlist[errno]);
		}
	} else if (pid > 0) {
		DBG("forked; cleaning up parent %d\n", getpid());
		delete_debug_socket(getpid());
	}

	return pid;
}

#ifdef INTERPOSE_EXIT
/* libc interposer */
void exit(int status) {
	delete_debug_socket();
	DBG("exit(%d)\n", status);
	libc_exit(status);
}
#endif

#ifdef INTERPOSE__EXIT
/* libc interposer */
void _exit(int status) {
	delete_debug_socket();
	DBG("_exit(%d)\n", status);
	libc__exit(status);
}
#endif

sighandler_t signal(int signum, sighandler_t handler)
{
	DBG("signal #%d registered\n", signum);
	return libc_signal(signum, handler);
}

int pthread_sigmask(int how, const sigset_t *set, sigset_t *oldset)
{
#if 0
	sigdelset((sigset_t *)set, SIGTERM);
	sigdelset((sigset_t *)set, SIGINT);
	DBG("interposing signal #%d\n", SIGTERM);
	DBG("interposing signal #%d\n", SIGINT);
#endif
	assert(pthread_pthread_sigmask != NULL);
	return pthread_pthread_sigmask(how, set, oldset);
}
