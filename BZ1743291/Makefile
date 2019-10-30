CFLAGS = -g -shared -fPIC

all: liballoc.so

liballoc.so: alloc.o
	$(LINK.c) -o $@ $^ -ldl

test: liballoc.so
	LD_PRELOAD=./$^ ls

*.o *.c: Makefile

clean:
	$(RM) lib*.o *.o
