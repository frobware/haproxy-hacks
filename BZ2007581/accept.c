/* accept4/close interceptor */

#define _GNU_SOURCE
#include <stdio.h>
#include <stdlib.h>
#include <dlfcn.h>
#include <time.h>
#include <stdarg.h>
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
static sighandler_t (*real_signal)(int signum, sighandler_t handler);

struct intercepted_accept {
	time_t accept_time;
	char address[INET6_ADDRSTRLEN];
	int port;
	int family;
	int fd;			/* sanity check */
};

static struct intercepted_accept fds[65536*2];

static pthread_mutex_t fds_lock;
static pthread_t dump_handler_tid;
static char *socket_path;

#define LOCK_FDS						\
	do {							\
		if (pthread_mutex_lock(&fds_lock) == -1) {	\
			perror("pthread_mutex_lock");		\
		}						\
	} while (0)

#define UNLOCK_FDS						\
	do {							\
		if (pthread_mutex_unlock(&fds_lock) == -1) {	\
			perror("pthread_mutex_unlock");		\
		}						\
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

static void teardown(void)
{
	if (socket_path != NULL && *socket_path) {
		unlink(socket_path);
	}
}

static void *dump_handler(void *userarg)
{
	struct sockaddr_un addr;
	int fd;

	if (atexit(teardown) != 0) {
		perror("atexit");
		exit(EXIT_FAILURE);
	}

	if (socket_path == NULL) {
		perror("no socket path");
		return NULL;
	}

	if ((fd = socket(AF_UNIX, SOCK_STREAM, 0)) == -1) {
		perror("socket error");
		return NULL;
	}

	memset(&addr, 0, sizeof(addr));
	addr.sun_family = AF_UNIX;
	strncpy(addr.sun_path, socket_path, sizeof(addr.sun_path)-1);
	unlink(socket_path);

	fprintf(stderr, "dumper socket: %s\n", addr.sun_path);

	if (bind(fd, (struct sockaddr*)&addr, sizeof(addr)) == -1) {
		perror("bind error");
		return NULL;
	}

	if (listen(fd, 5) == -1) {
		perror("listen error");
		exit(-1);
	}

	while (1) {
		int cl;
		int rc;		/* result code */
		char buf[100];

		if ((cl = accept(fd, NULL, NULL)) == -1) {
			perror("accept error");
			continue;
		}

		rc = read(cl, buf, sizeof buf);
		if (rc == -1) {
			perror("read");
			close(cl);
		}

		char msg[8192];
		char tbuf[256] = { '\0' };
		struct tm tm_tmp;

		LOCK_FDS;
		for (size_t i = 0; i < NELEMENTS(fds); i++) {
			if (fds[i].fd != -1) {
				*tbuf = '\0';
				localtime_r(&fds[i].accept_time, &tm_tmp);
				strftime(tbuf, sizeof(tbuf), "%FT%T.%6S", &tm_tmp);
				int n = snprintf(msg, sizeof msg, "%s [haproxy:%d] %d %s%s%s:%d\n",
						 tbuf,
						 getpid(),
						 fds[i].fd,
						 fds[i].family == AF_INET6 ? "[" : "",
						 fds[i].address,
						 fds[i].family == AF_INET6 ? "]" : "",
						 fds[i].port);
				if (n > 0) {
					printf("WRITE: %zd\n", write(cl, msg, n));
				}
			}
		}
		UNLOCK_FDS;
		close(cl);
	}

	close(fd);
}

int accept4(int sockfd, struct sockaddr *sa, socklen_t *addrlen, int flags)
{
	int fd = real_accept4(sockfd, sa, addrlen, flags);

	fprintf(stderr, "INTERCEPTED %d accept4\n", fd);

	if (fd == -1) {
		return -1;
	}

	LOCK_FDS;


	switch(sa->sa_family) {
	case AF_INET:
		fds[fd].fd = fd;
		fds[fd].accept_time = time(NULL);
		inet_ntop(AF_INET, &(((struct sockaddr_in *)sa)->sin_addr), fds[fd].address, sizeof fds[fd].address);
		fds[fd].port = ((struct sockaddr_in *)sa)->sin_port;
		fds[fd].family = sa->sa_family;
		break;
	case AF_INET6:
		fds[fd].fd = fd;
		fds[fd].accept_time = time(NULL);
		inet_ntop(AF_INET6, &(((struct sockaddr_in6 *)sa)->sin6_addr), fds[fd].address, sizeof fds[fd].address);
		fds[fd].port = ((struct sockaddr_in *)sa)->sin_port;
		fds[fd].family = sa->sa_family;
		break;
	}

	UNLOCK_FDS;

	return fd;
}

sighandler_t signal(int signum, sighandler_t handler)
{
	fprintf(stderr, "SIGNAL %d\n", signum);
	return real_signal(signum, handler);
}

int close(int fd)
{
	assert(real_close != NULL);

	LOCK_FDS;
	if (fds[fd].fd == fd) {;
		fprintf(stderr, "INTERCEPTED %d close\n", fd);
		fds[fd].fd = -1; /* clear */
	}
	UNLOCK_FDS;

	return real_close(fd);
}

// printf GET | socat UNIX-CONNECT:/tmp/server.sock -

static void sig_handler(int signum)
{
	if (signum == SIGTERM || signum == SIGINT) {
		teardown();
		/* restore default handler */
		signal(signum, SIG_DFL);
		// re-raise for default behaviour.
		kill(getpid(), signum);
	}
}

static void __attribute__((constructor)) setup(void)
{
	fprintf(stderr, "INTERCEPT constructor\n");

	socket_path = mprintf("/tmp/haproxy.sock.connections", getpid());
	if (socket_path == NULL) {
		perror("malloc");
		exit(EXIT_FAILURE);
	}

	LOCK_FDS;
	for (size_t i = 0; i < NELEMENTS(fds); i++) {
		fds[i].fd = -1;	/* cleared */
	}
	UNLOCK_FDS;

	if ((real_close = dlsym(RTLD_NEXT, "close")) == NULL) {
		perror("dlsym close");
		exit(EXIT_FAILURE);
	}

	if ((real_accept4 = dlsym(RTLD_NEXT, "accept4")) == NULL) {
		perror("dlsym accept4");
		exit(EXIT_FAILURE);
	}

	if ((real_signal = dlsym(RTLD_NEXT, "signal")) == NULL) {
		perror("dlsym signal");
		exit(EXIT_FAILURE);
	}

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
	if (signal(SIGTERM, sig_handler) == SIG_ERR) {
		perror("SIGTERM registration failed");
		exit(EXIT_FAILURE);
	}
	if (signal(SIGINT, sig_handler) == SIG_ERR) {
		perror("SIGINT registration failed");
		exit(EXIT_FAILURE);
	}
}
