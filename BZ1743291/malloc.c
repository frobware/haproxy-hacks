/* malloc interceptor; adjust size to balloon allocations. Not MT-Safe. */

#define _GNU_SOURCE
#include <stdio.h>
#include <dlfcn.h>
#include <assert.h>

static void *(*real_malloc)(size_t size);

static void malloc_init(void)
{
	/* MT-Safe? Nope. */
	real_malloc = dlsym(RTLD_NEXT, "malloc");
	assert(real_malloc);
}

#define ONE_MB 1048576		/* 1<<20 */

#ifndef MALLOC_BALLOON
#define MALLOC_BALLOON 10 * ONE_MB
#endif

void *malloc(size_t size)
{
	void *ptr;

	if (size > 0) {
	  size += MALLOC_BALLOON;
	}

        if (real_malloc == NULL) {
		malloc_init();
	}
	ptr = real_malloc(size);
#if 0
	fprintf(stderr, "malloc %zd == %p\n", size, ptr);
#endif
	return ptr;
}
