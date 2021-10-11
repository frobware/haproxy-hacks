/* malloc/calloc/free interceptor; adjust size to balloon allocations. */

#define _GNU_SOURCE
#include <dlfcn.h>
#include <memory.h>
#include <stdio.h>
#include <stdlib.h>
#include <errno.h>
#include <unistd.h>
#include <sys/types.h>
#include <assert.h>

#define ONE_MB 1048576 /* 1<<20 */

#ifndef MALLOC_BALLOON
#define MALLOC_BALLOON ONE_MB * 10
#endif

#ifndef CALLOC_BALLOON
#define CALLOC_BALLOON 0
#endif

#define MALLOC_INTERPOSER_DEBUG

#ifdef MALLOC_INTERPOSER_DEBUG
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

static void *(*libc_malloc)(size_t size);
static void *(*libc_calloc)(size_t n, size_t size);
static void *(*libc_free)(void *ptr);

static int malloc_balloon = MALLOC_BALLOON;
static int calloc_balloon = CALLOC_BALLOON;

static size_t calloc_buffer_alloc = 0;
static unsigned char calloc_buffer[1<<20];

static FILE *fp_dbg;

void *malloc(size_t size) {
	void *ptr;

	assert(libc_malloc != NULL);
	assert(libc_calloc != NULL);

	if (size > 0) {
		size += malloc_balloon;
	}

	ptr = libc_malloc(size);
	DBG("size %zd ptr %p\n", size, ptr);
	return ptr;
}

void *calloc(size_t n, size_t size) {
	void *ptr;

	if (libc_calloc == NULL) {
		/* a one time only affair because dlsym calls
		 * calloc(yay!). Force the first call to calloc to
		 * come from a static buffer. */
		calloc_buffer_alloc++;
		assert(calloc_buffer_alloc == 1);
		return calloc_buffer;
	}

	if (size > 0) {
		size += calloc_balloon;
	}

	assert(libc_calloc != NULL);
	ptr = libc_calloc(n, size);
	DBG("n=%10zd size=%8zd ptr %p\n", n, size, ptr);
#if 0
	fprintf(stderr, "calloc n=%10zd size=%8zd == %p\n", n, size, ptr);
#endif
	return ptr;
}

void free(void *p) {
	if ((unsigned char *)p == calloc_buffer) {
		fprintf(stderr, "ignoring free in calloc buffer\n");
		return;
	}
	DBG("ptr %p\n", p);
	assert(libc_free != NULL);
	libc_free(p);
}

static __attribute__((constructor (101))) void setup(void)
{
	fp_dbg = stderr;

	if (libc_malloc == NULL) {
		if ((libc_malloc = dlsym(RTLD_NEXT, "malloc")) == NULL) {
			PERROR("error: dlsym(malloc): %s\n", strerror_r(errno, _errbuf, sizeof(_errbuf)));
			_exit(EXIT_FAILURE); /* has to be fatal */
		}
	}

	if (libc_free == NULL) {
		if ((libc_free = dlsym(RTLD_NEXT, "free")) == NULL) {
			PERROR("error: dlsym(free): %s\n", strerror_r(errno, _errbuf, sizeof(_errbuf)));
			_exit(EXIT_FAILURE); /* has to be fatal */
		}
	}

	if (libc_calloc == NULL) {
		if ((libc_calloc = dlsym(RTLD_NEXT, "calloc")) == NULL) {
			PERROR("error: dlsym(calloc): %s\n", strerror_r(errno, _errbuf, sizeof(_errbuf)));
			_exit(EXIT_FAILURE); /* has to be fatal */
		}
	}


	if (libc_malloc == NULL) {
		if ((libc_malloc = dlsym(RTLD_NEXT, "malloc")) == NULL) {
			PERROR("error: dlsym(malloc): %s\n", strerror_r(errno, _errbuf, sizeof(_errbuf)));
			_exit(EXIT_FAILURE); /* has to be fatal */
		}
	}

	DBG("MALLOC-INTERPOSER initialised\n");
	/* disabled post initialisation */
	fp_dbg = NULL;
	return;
}
