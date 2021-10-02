CFLAGS = -g -shared -fPIC -Werror -Wall

all: libaccept.so

libaccept.so: accept.o
	$(LINK.c) -o $@ $^ -ldl -lpthread

test: libaccept.so
	LD_PRELOAD=./$^ ls

*.o *.c: Makefile

clean:
	$(RM) lib*.o *.o
