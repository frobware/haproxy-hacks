/* accept4/close interceptor */

#define _GNU_SOURCE
#include <stdio.h>
#include <stdarg.h>
#include <stdlib.h>
#include <time.h>
#include <dlfcn.h>
#include <sys/time.h>
#include <unistd.h>
#include <signal.h>
#include <sys/types.h>
#include <sys/socket.h>
#include <sys/un.h>
#include <arpa/inet.h>
#include <pthread.h>
#include <limits.h>

#undef NDEBUG			/* always want an assert to fire */
#include <assert.h>

#define NELEMENTS(A) ((sizeof(A) / sizeof(A[0])))

static int (*real_accept4)(int sockfd, struct sockaddr *addr, socklen_t *addrlen, int flags);
static int (*real_close)(int fd);

struct intercepted_accept {
	struct timespec accept_time;
	char address[INET6_ADDRSTRLEN];
	int port;
	int family;
	int fd;			/* sanity check */
};


static pthread_mutex_t accepted_fds_lock;
static struct intercepted_accept accepted_fds[65536*2];

static char socket_path[PATH_MAX];
static int initialised;

#define LOCK_ACCEPTED_FDS						\
	do {								\
		if (pthread_mutex_lock(&accepted_fds_lock) == -1) {	\
			perror("pthread_mutex_lock");			\
		}							\
	} while (0)

#define UNLOCK_ACCEPTED_FDS						\
	do {								\
		if (pthread_mutex_unlock(&accepted_fds_lock) == -1) {	\
			perror("pthread_mutex_unlock");			\
		}							\
	} while (0)


/* Return a - b */
static void timespec_diff(const struct timespec *a, const struct timespec *b, struct timespec *result) {
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

	return snprintf(s, sz, "%3d day%s %2d hour%s %2d minute%s %2d sec%s",
			days,
			days > 0 ? "s" : "",
			hours,
			hours > 0 ? "s" : "",
			mins,
			mins > 0 ? "s" : "",
			secs,
			secs > 0 ? "s" : "");
}

#if 0
static char *write_inetaddr_as_str(const struct sockaddr *sa, char *s, size_t maxlen, int *port)
{
	switch(sa->sa_family) {
	case AF_INET:
		inet_ntop(AF_INET, &(((struct sockaddr_in *)sa)->sin_addr), s, maxlen);
		*port = ((struct sockaddr_in *)sa)->sin_port;
		break;

	case AF_INET6:
		inet_ntop(AF_INET6, &(((struct sockaddr_in6 *)sa)->sin6_addr), s, maxlen);
		*port = ((struct sockaddr_in6 *)sa)->sin6_port;
		break;
	default:
		strncpy(s, "Unknown AF", maxlen);
		return NULL;
	}

	return s;
}
#endif

static int write_connection_state(int fd)
{
	char msg[8192];
	struct timespec elapsed_time;

	LOCK_ACCEPTED_FDS;
	for (size_t i = 0; i < NELEMENTS(accepted_fds); i++) {
		if (accepted_fds[i].fd == -1) {
			continue;
		}

		if (!(accepted_fds[i].family == AF_INET ||
		      accepted_fds[i].family == AF_INET6)) {
			continue;
		}

		char timebuf[256] = { '\0' };
		struct timespec current_time = { 0, 0 };

		clock_gettime(CLOCK_MONOTONIC, &current_time);
		timespec_diff(&current_time, &accepted_fds[i].accept_time, &elapsed_time);
		human_time(&elapsed_time, timebuf, sizeof(timebuf));

		int n = snprintf(msg, sizeof(msg), "pid:%d %d %s%s%s %d %s\n",
				 getpid(),
				 accepted_fds[i].fd,
				 accepted_fds[i].family == AF_INET6 ? "[" : "",
				 accepted_fds[i].address,
				 accepted_fds[i].family == AF_INET6 ? "]" : "",
				 accepted_fds[i].port,
				 timebuf);
		if (n > 0) {
			int wrote = write(fd, msg, n);
			if (wrote != n) {
				return -1;
			}
		}
	}
	UNLOCK_ACCEPTED_FDS;

	return 0;
}

static void *dump_handler(void *userarg)
{
	if (!initialised) {
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

	while (1) {
		int cl;

		if ((cl = accept(listen_fd, NULL, NULL)) == -1) {
			perror("accept error");
			continue;
		}

		if (write_connection_state(cl) == -1) {
			perror("write");
		}

		real_close(cl);
	}

	fprintf(stderr, "**** ACCEPT-INTERPOSER: dumper exiting\n");
	pthread_exit(NULL);
	return NULL;
}

static void sig_handler(int signum)
{
	if (*socket_path != '\0') {
		fprintf(stderr, "**** ACCEPT-INTERPOSER unlinking %s\n", socket_path);
		if (unlink(socket_path) != 0) {
			perror("unlink");
		}
	}

	/* restore default handler. */
	signal(signum, SIG_DFL);

	/* re-raise this signal for default behaviour. */
	if (kill(getpid(), signum) != 0) {
		exit(0);	/* last gasp saloon. */
	}
}

/* libc interposer */
int accept4(int sockfd, struct sockaddr *sa, socklen_t *addrlen, int flags)
{
	assert(real_accept4 != NULL);

	int fd = real_accept4(sockfd, sa, addrlen, flags);

	if (!initialised) {
		return fd;
	}

	fprintf(stderr, "**** ACCEPT-INTERPOSER accept4() = %d\n", fd);

	if ((sa == NULL || addrlen == 0) ||
	    (fd < 0 || fd >= NELEMENTS(accepted_fds))) {
		return fd;
	}

	LOCK_ACCEPTED_FDS;
	assert(accepted_fds[fd].fd == -1); /* sanity check; set in close() */
	accepted_fds[fd].family = sa->sa_family;

	switch(sa->sa_family) {
	case AF_INET:
		accepted_fds[fd].fd = fd;
		clock_gettime(CLOCK_MONOTONIC, &accepted_fds[fd].accept_time);
		inet_ntop(AF_INET, &(((struct sockaddr_in *)sa)->sin_addr), accepted_fds[fd].address, sizeof accepted_fds[fd].address);
		accepted_fds[fd].port = ((struct sockaddr_in *)sa)->sin_port;
		break;
	case AF_INET6:
		accepted_fds[fd].fd = fd;
		clock_gettime(CLOCK_MONOTONIC, &accepted_fds[fd].accept_time);
		inet_ntop(AF_INET6, &(((struct sockaddr_in6 *)sa)->sin6_addr), accepted_fds[fd].address, sizeof accepted_fds[fd].address);
		accepted_fds[fd].port = ((struct sockaddr_in *)sa)->sin_port;
		break;
	}

	UNLOCK_ACCEPTED_FDS;

	return fd;
}

/* libc interposer */
int close(int fd)
{
	assert(real_close != NULL);
	int rc = real_close(fd);

	if (initialised) {
		if (fd >= 0 && fd <= NELEMENTS(accepted_fds)) {
			LOCK_ACCEPTED_FDS;
			accepted_fds[fd].fd = -1; /* clear */
			UNLOCK_ACCEPTED_FDS;
		}
	}

	fprintf(stderr, "**** ACCEPT-INTERPOSER close(%d) = %d\n", fd, rc);

	return rc;
}

static void __attribute__((destructor)) teardown(void) {
	if (*socket_path != '\0') {
		fprintf(stderr, "**** ACCEPT-INTERPOSER unlinking %s\n", socket_path);
		if (unlink(socket_path) != 0) {
			perror("unlink");
		}
	}
}

static void __attribute__((constructor)) setup(void)
{
	pthread_t dump_handler_tid;

	if ((real_accept4 = dlsym(RTLD_NEXT, "accept4")) == NULL) {
		perror("dlsym accept4 error; cannot continue");
		exit(EXIT_FAILURE);
	}

	if ((real_close = dlsym(RTLD_NEXT, "close")) == NULL) {
		perror("dlsym close error; cannot continue");
		exit(EXIT_FAILURE);
	}

	if (snprintf(socket_path, sizeof(socket_path), "/tmp/haproxy-%d.connections", getpid()) < 1) {
		return;
	}

	if (atexit(teardown) != 0) {
		perror("atexit");
		return;
	}

	LOCK_ACCEPTED_FDS;
	for (size_t i = 0; i < NELEMENTS(accepted_fds); i++) {
		accepted_fds[i].fd = -1;	/* cleared */
	}
	UNLOCK_ACCEPTED_FDS;

	if (pthread_create(&dump_handler_tid, NULL, &dump_handler, NULL) != 0) {
		perror("pthread_create");
		return;
	}

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
	 * We register signal handlers so that we cleanup the socket
	 * path that is created.
	 */
	if (signal(SIGTERM, sig_handler) == SIG_ERR) {
		perror("signal handler for SIGTERM registration failed");
		exit(EXIT_FAILURE);
	}

	if (signal(SIGINT, sig_handler) == SIG_ERR) {
		perror("signal handler for SIGINT registration failed");
		exit(EXIT_FAILURE);
	}

	initialised = 1;
	fprintf(stderr, "**** ACCEPT-INTERPOSER is initialised\n");
	return;
}
