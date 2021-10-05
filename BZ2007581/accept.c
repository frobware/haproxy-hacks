/* accept4/close interceptor */

#define _GNU_SOURCE
#include <stdio.h>
#include <stdarg.h>
#include <stdlib.h>
#include <limits.h>
#include <signal.h>
#include <string.h>
#include <time.h>
#include <dlfcn.h>
#include <sys/time.h>
#include <unistd.h>
#include <sys/types.h>
#include <sys/socket.h>
#include <sys/un.h>
#include <arpa/inet.h>
#include <pthread.h>

#undef NDEBUG			/* always want an assert to fire */
#include <assert.h>

#define NELEMENTS(A) ((sizeof(A) / sizeof(A[0])))

static int (*real_accept)(int sockfd, struct sockaddr *addr, socklen_t *addrlen);
static int (*real_accept4)(int sockfd, struct sockaddr *addr, socklen_t *addrlen, int flags);
static int (*real_close)(int fd);

struct intercepted_accept {
	struct timespec accept_time;
	char local_address[INET6_ADDRSTRLEN];
	int local_port;
	int family;
	int fd;			/* sanity check */
	char peer_address[INET6_ADDRSTRLEN];
	int peer_port;
};

static pthread_mutex_t fdtab_lock;
static struct intercepted_accept fdtab[65536*2];

static char socket_path[PATH_MAX];
static volatile int initialised;

#define LOCK_FDTAB						\
	do {							\
		if (pthread_mutex_lock(&fdtab_lock) == -1) {	\
			perror("pthread_mutex_lock");		\
		}						\
	} while (0)

#define UNLOCK_FDTAB						\
	do {							\
		if (pthread_mutex_unlock(&fdtab_lock) == -1) {	\
			perror("pthread_mutex_unlock");		\
		}						\
	} while (0)


/* Return a - b */
static void timespec_sub(const struct timespec *a, const struct timespec *b, struct timespec *result) {
	result->tv_sec  = a->tv_sec  - b->tv_sec;
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

	return snprintf(s, sz, "%d day%s, %2d hour%s, %2d minute%s, %2d sec%s",
			days,
			days > 0 ? "s" : "s",
			hours,
			hours > 0 ? "s" : "s",
			mins,
			mins > 0 ? "s" : "s",
			secs,
			secs > 0 ? "s" : "s");
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

static int write_connection_state_locked(int fd)
{
	char msg[8192];
	struct timespec elapsed_time;

	LOCK_FDTAB;
	for (size_t i = 0; i < NELEMENTS(fdtab); i++) {
		if (fdtab[i].fd == -1) {
			continue;
		}

		if (!(fdtab[i].family == AF_INET ||
		      fdtab[i].family == AF_INET6)) {
			continue;
		}

		char timebuf[256] = { '\0' };
		struct timespec current_time = { 0, 0 };

		clock_gettime(CLOCK_MONOTONIC, &current_time);
		timespec_sub(&current_time, &fdtab[i].accept_time, &elapsed_time);
		human_time(&elapsed_time, timebuf, sizeof(timebuf));

		int n = snprintf(msg, sizeof(msg), "pid:%d fd:%d %s%s%s:%d %s%s%s:%d %ld secs (%s)\n",
				 getpid(),
				 fdtab[i].fd,
				 fdtab[i].family == AF_INET6 ? "[" : "",
				 fdtab[i].local_address,
				 fdtab[i].family == AF_INET6 ? "]" : "",
				 fdtab[i].local_port,
				 fdtab[i].family == AF_INET6 ? "[" : "",
				 fdtab[i].peer_address,
				 fdtab[i].family == AF_INET6 ? "]" : "",
				 fdtab[i].peer_port,
				 elapsed_time.tv_sec,
				 timebuf);
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

static void *connection_state_handler(void *userarg)
{
	if (!initialised) {
		assert(0);
		return NULL;
	}

	int listen_fd;
	struct sockaddr_un addr;

	if ((listen_fd = socket(AF_UNIX, SOCK_STREAM|SOCK_CLOEXEC, 0)) == -1) {
		perror("socket error");
		return NULL;
	}

	memset(&addr, 0, sizeof(addr));
	addr.sun_family = AF_UNIX;
	strncpy(addr.sun_path, socket_path, sizeof(addr.sun_path)-1);
	unlink(addr.sun_path);

	if (bind(listen_fd, (struct sockaddr *)&addr, sizeof(addr)) == -1) {
		perror("bind error");
		return NULL;
	}

	if (listen(listen_fd, 5) == -1) {
		perror("listen error");
		return NULL;
	}

	fprintf(stdout, "[%s:%d] **** ACCEPT-INTERPOSER listening at: %s\n", "haproxy", getpid(), addr.sun_path);

	while (1) {
		int cl;

		if ((cl = real_accept(listen_fd, NULL, NULL)) == -1) {
			perror("accept error");
			continue;
		}

		fprintf(stdout, "[%s:%d] **** ACCEPT-INTERPOSER new debug connection: %d\n", "haproxy", getpid(), cl);
		
		if (write_connection_state_locked(cl) == -1) {
			perror("write");
		}

		real_close(cl);
	}

	fprintf(stderr, "[%s:%d] **** ACCEPT-INTERPOSER: connection_state_handler exited\n", "haproxy", getpid());
	pthread_exit(NULL);
	return NULL;
}

static void signal_handler(int signum)
{
	if (*socket_path != '\0') {
		fprintf(stderr, "[%s:%d] **** ACCEPT-INTERPOSER unlinking %s\n", "haproxy", getpid(), socket_path);
		if (unlink(socket_path) != 0) {
			perror("unlink");
		}
	}

	/* restore default handler. */
	signal(signum, SIG_DFL);

	/* re-raise this signal for default behaviour. */
	if (kill(getpid(), signum) != 0) {
		exit(0);
	}
}

/* libc interposer */
int accept(int sockfd, struct sockaddr *sa, socklen_t *salen)
{
	assert(real_accept != NULL);
	return accept4(sockfd, sa, salen, 0);
}

/* libc interposer */
int accept4(int sockfd, struct sockaddr *sa, socklen_t *salen, int flags)
{
	assert(real_accept4 != NULL);

	int clientfd = real_accept4(sockfd, sa, salen, flags);

	if (!initialised || clientfd == -1) {
		return clientfd;
	}

	fprintf(stderr, "[%s:%d] **** ACCEPT-INTERPOSER accept4() = %d\n", "haproxy", getpid(), clientfd);

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

	switch(sa->sa_family) {
	case AF_INET:
		fdtab[clientfd].family = sa->sa_family;
		fdtab[clientfd].fd = clientfd;
		if (getsockname(sockfd, &local_addr, &local_addrlen) == 0) {
			fdtab[clientfd].local_port = ntohs(((struct sockaddr_in *)&local_addr)->sin_port);
			inet_ntop(AF_INET, &(((struct sockaddr_in *)&local_addr)->sin_addr), fdtab[clientfd].local_address, sizeof(fdtab[clientfd].local_address));
		}
                fdtab[clientfd].peer_port = ntohs(((struct sockaddr_in *)sa)->sin_port);
		inet_ntop(sa->sa_family, &((struct sockaddr_in *)sa)->sin_addr, fdtab[clientfd].peer_address, sizeof(fdtab[clientfd].peer_address));
                printf("%s %d\n", fdtab[clientfd].peer_address, fdtab[clientfd].peer_port);
                printf("%s %d\n", fdtab[clientfd].local_address, fdtab[clientfd].local_port);
		break;
	case AF_INET6:
		fdtab[clientfd].family = sa->sa_family;
		fdtab[clientfd].fd = clientfd;
		if (getsockname(sockfd, &local_addr, &local_addrlen) == 0) {
			fdtab[clientfd].local_port = ntohs(((struct sockaddr_in6 *)&local_addr)->sin6_port);
			inet_ntop(AF_INET, &(((struct sockaddr_in6 *)&local_addr)->sin6_addr), fdtab[clientfd].local_address, sizeof(fdtab[clientfd].local_address));
		}
                fdtab[clientfd].peer_port = ntohs(((struct sockaddr_in6 *)sa)->sin6_port);
		inet_ntop(sa->sa_family, &((struct sockaddr_in6 *)sa)->sin6_addr, fdtab[clientfd].peer_address, sizeof(fdtab[clientfd].peer_address));
                printf("%s %d\n", fdtab[clientfd].peer_address, fdtab[clientfd].peer_port);
                printf("%s %d\n", fdtab[clientfd].local_address, fdtab[clientfd].local_port);
		break;
	}

	UNLOCK_FDTAB;

	return clientfd;
}

/* libc interposer */
int close(int fd)
{
	assert(real_close != NULL);

	if (initialised) {
		if (fd >= 0 && fd <= NELEMENTS(fdtab)) {
			LOCK_FDTAB;
			fdtab[fd].fd = -1; /* clear */
			UNLOCK_FDTAB;
		}
	}

	int rc = real_close(fd);
	fprintf(stderr, "[%s:%d] **** ACCEPT-INTERPOSER close(%d) = %d\n", "haproxy", getpid(), fd, rc);
	return rc;
}

static void __attribute__((destructor)) teardown(void) {
	if (*socket_path != '\0') {
		fprintf(stderr, "[%s:%d] **** ACCEPT-INTERPOSER unlinking %s\n", "haproxy", getpid(), socket_path);
		if (unlink(socket_path) != 0) {
			perror("unlink");
		}
	}
}

static void __attribute__((constructor (101))) setup(void)
{
	pthread_t connection_state_tid;

	snprintf(socket_path, sizeof(socket_path), "/tmp/haproxy-%d.connections", getpid());

	if ((real_accept4 = dlsym(RTLD_NEXT, "accept4")) == NULL) {
		perror("dlsym accept4 error; cannot continue");
		exit(EXIT_FAILURE); /* has to be fatal */
	}

	if ((real_accept = dlsym(RTLD_NEXT, "accept")) == NULL) {
		perror("dlsym accept error; cannot continue");
		exit(EXIT_FAILURE); /* has to be fatal */
	}

	if ((real_close = dlsym(RTLD_NEXT, "close")) == NULL) {
		perror("dlsym close error; cannot continue");
		exit(EXIT_FAILURE); /* has to be fatal */
	}

	if (atexit(teardown) != 0) {
		perror("atexit");
		return;
	}

	LOCK_FDTAB;
	for (size_t i = 0; i < NELEMENTS(fdtab); i++) {
		fdtab[i].fd = -1;	/* cleared */
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
	 * We register SIGTERM and SIGINT so that we cleanup the
	 * socket path that is created.
	 */
	if (signal(SIGTERM, signal_handler) == SIG_ERR) {
		perror("signal handler for SIGTERM registration failed");
		return;
	}

	if (signal(SIGINT, signal_handler) == SIG_ERR) {
		perror("signal handler for SIGINT registration failed");
		return;
	}

	initialised = 1;

	if (pthread_create(&connection_state_tid, NULL, &connection_state_handler, NULL) != 0) {
		perror("pthread_create");
		return;
	}

	fprintf(stderr, "[%s:%d] **** ACCEPT-INTERPOSER is initialised\n", "haproxy", getpid());
	return;
}
