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

#undef NDEBUG			/* always want an assert to fire */
#include <assert.h>

#define NELEMENTS(A) ((sizeof(A) / sizeof(A[0])))

static int (*real_accept4)(int sockfd, struct sockaddr *addr, socklen_t *addrlen, int flags);
static int (*real_close)(int fd);

#if 0
static sighandler_t (*real_signal)(int signum, sighandler_t handler);
#endif

struct intercepted_accept {
	struct timespec accept_time;
	char address[INET6_ADDRSTRLEN];
	int port;
	int family;
	int fd;			/* sanity check */
};

static struct intercepted_accept accepted_fds[65536*2];

static pthread_mutex_t accepted_fds_lock;
static pthread_t dump_handler_tid;
static const char *socket_path;
static int dumper_listen_fd;
static int terminate;

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


static char *mprintf(const char *fmt, ...)
{
	size_t size = 0;
	char *p = NULL;
	va_list ap;

	/* Determine required size. */
	va_start(ap, fmt);
	size = vsnprintf(p, size, fmt, ap);
	va_end(ap);

	size++; /* For '\0' */

	if ((p = malloc(size)) == NULL) {
		return NULL;
	}

	va_start(ap, fmt);
	size = vsnprintf(p, size, fmt, ap);
	va_end(ap);

	return p;
}

/* Return a - b */
static void timespec_diff(const struct timespec *a, const struct timespec *b, struct timespec *result) {
	result->tv_sec  = a->tv_sec  - b->tv_sec;
	result->tv_nsec = a->tv_nsec - b->tv_nsec;
	if (result->tv_nsec < 0) {
		--result->tv_sec;
		result->tv_nsec += 1000000000L;
	}
}

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
		if (!(accepted_fds[i].family == AF_INET || accepted_fds[i].family == AF_INET6)) {
			continue;
		}

		char timebuf[256] = { '\0' };
		struct timespec current_time = { 0, 0 };

		clock_gettime(CLOCK_MONOTONIC, &current_time);
		timespec_diff(&current_time, &accepted_fds[i].accept_time, &elapsed_time);
		human_time(&elapsed_time, timebuf, sizeof(timebuf));
		
		int n = snprintf(msg, sizeof(msg), "[haproxy:%d] %d %s%s%s %d %s\n",
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
	struct sockaddr_un addr;

	if (socket_path == NULL) {
		perror("no socket path");
		return NULL;
	}

	if ((dumper_listen_fd = socket(AF_UNIX, SOCK_STREAM|SOCK_CLOEXEC, 0)) == -1) {
		perror("socket error");
		return NULL;
	}

	memset(&addr, 0, sizeof(addr));
	addr.sun_family = AF_UNIX;
	strncpy(addr.sun_path, socket_path, sizeof(addr.sun_path)-1);
	unlink(socket_path);

	fprintf(stderr, "dumper socket: %s\n", addr.sun_path);

	if (bind(dumper_listen_fd, (struct sockaddr *)&addr, sizeof(addr)) == -1) {
		perror("bind error");
		return NULL;
	}

	if (listen(dumper_listen_fd, 5) == -1) {
		perror("listen error");
		exit(EXIT_FAILURE);
	}

	while (!terminate) {
		int cl;
#if 0
		char buf[100];
#endif

		if ((cl = accept4(dumper_listen_fd, NULL, NULL, 0)) == -1) {
			if (terminate) {
				return NULL;
			}
			perror("accept error");
			continue;
		}

		fprintf(stderr, "new debug client: %d\n", cl);

		if (terminate) {
			close(cl);
			return NULL;
		}

		if (write_connection_state(cl) == -1) {
			perror("write");
		}
#if 0
		if (read(cl, buf, sizeof(buf)) == -1) {
			perror("read");
		} else {
			if (write_connection_state(cl) == -1) {
				perror("write");
			}
		}
#endif
		close(cl);
	}

	if (socket_path != NULL && *socket_path) {
		fprintf(stderr, "unlinking %s\n", socket_path);
		unlink(socket_path);
	}

	fprintf(stderr, "INTERCEPTED: dumper exiting\n");
	pthread_exit(NULL);
	return NULL;
}

#if 0
// This exists to verify that HAProxy doesn't register a signal
// handler for either SIGTERM or SIGINT. This shim uses TERM and INT
// to clean up the socket pathname it creates. Uncomment this to
// verify that the signum's that HAProxy registers do not include
// SIGTERM or SIGINT.
sighandler_t signal(int signum, sighandler_t handler)
{
	fprintf(stderr, "SIGNAL %d\n", signum);
	return real_signal(signum, handler);
}
#endif

static void __attribute__((destructor)) teardown(void) {
	fprintf(stderr, "INTERCEPT destructor\n");

	if (socket_path != NULL && *socket_path) {
		fprintf(stderr, "unlinking %s\n", socket_path);
		unlink(socket_path);
	}
}

static void sig_handler(int signum)
{
	fprintf(stderr, "SIGNAL %d\n", signum);

	terminate = 1;
#if 1
	int fd;

	if ((fd = socket(AF_UNIX, SOCK_STREAM, 0)) == -1) {
		perror("socket error");
		exit(-1);
	}

	struct sockaddr_un addr;
	memset(&addr, 0, sizeof(addr));
	addr.sun_family = AF_UNIX;
	strncpy(addr.sun_path, socket_path, sizeof(addr.sun_path)-1);

	if (connect(fd, (struct sockaddr *)&addr, sizeof(addr)) == -1) {
		perror("connect error");
		exit(-1);
	}
#endif

	/* close(dumper_listen_fd); */
	void *retval = NULL;
	fprintf(stderr, "Waiting for thread exit\n");
	pthread_join(dump_handler_tid, &retval);
	fprintf(stderr, "Waiting for thread has exited\n");

	/* remove the named socket */
	teardown();

	/* restore default handler. */
	signal(signum, SIG_DFL);

	/* re-raise this signal for default behaviour. */
	kill(getpid(), signum);
}

static void __attribute__((constructor)) setup(void)
{
	long a;
	int i;

	a = sysconf(_SC_ATEXIT_MAX);
	printf("ATEXIT_MAX = %ld\n", a);

	i = atexit(teardown);
	if (i != 0) {
		fprintf(stderr, "cannot set exit function\n");
		exit(EXIT_FAILURE);
	}

	fprintf(stderr, "INTERCEPT constructor\n");

	socket_path = mprintf("/tmp/haproxy.sock.connections", getpid());
	if (socket_path == NULL) {
		perror("malloc");
		exit(EXIT_FAILURE);
	}

	LOCK_ACCEPTED_FDS;
	for (size_t i = 0; i < NELEMENTS(accepted_fds); i++) {
		accepted_fds[i].fd = -1;	/* cleared */
	}
	UNLOCK_ACCEPTED_FDS;

	if ((real_accept4 = dlsym(RTLD_NEXT, "accept4")) == NULL) {
		perror("dlsym accept4");
		exit(EXIT_FAILURE);
	}

	if ((real_close = dlsym(RTLD_NEXT, "close")) == NULL) {
		perror("dlsym close");
		exit(EXIT_FAILURE);
	}

#if 0
	if ((real_signal = dlsym(RTLD_NEXT, "signal")) == NULL) {
		perror("dlsym signal");
		exit(EXIT_FAILURE);
	}
#endif

	if (pthread_create(&dump_handler_tid, NULL, &dump_handler, NULL) != 0) {
		perror("pthread_create");
		exit(EXIT_FAILURE);
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

	/* We register signal handlers so that we cleanup the socket
	 * path that is created. */
	if (signal(SIGTERM, sig_handler) == SIG_ERR) {
		perror("signal handler for SIGTERM registration failed");
		exit(EXIT_FAILURE);
	}
	if (signal(SIGINT, sig_handler) == SIG_ERR) {
		perror("signal handler for SIGINT registration failed");
		exit(EXIT_FAILURE);
	}
}

/* libc interposer */
int accept4(int sockfd, struct sockaddr *sa, socklen_t *addrlen, int flags)
{
	int fd = real_accept4(sockfd, sa, addrlen, flags);

	fprintf(stderr, "INTERCEPTED %d accept4\n", fd);

	if ((sa == NULL || addrlen == 0) ||
	    (fd < 0 || fd >= NELEMENTS(accepted_fds))) {
		return fd;	/* we have finite space. */
	}

	LOCK_ACCEPTED_FDS;
	assert(accepted_fds[fd].fd == -1);
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
	int rc;
	assert(real_close != NULL);
	LOCK_ACCEPTED_FDS;
	if (fd >= 0 && fd <= NELEMENTS(accepted_fds)) {
		if (accepted_fds[fd].fd == fd) {;
			fprintf(stderr, "INTERCEPTED %d closed\n", fd);
			accepted_fds[fd].fd = -1; /* clear */
		}
	}
	UNLOCK_ACCEPTED_FDS;
	rc = real_close(fd);
	fprintf(stderr, "INTERCEPTED %d close (%p) = %d\n", fd, real_close, rc);
	return rc;
}
