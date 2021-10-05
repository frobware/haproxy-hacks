CFLAGS = -g -ggdb3 -shared -fPIC -Wall -Werror -Wl,-init,init

#-fsanitize=address

all: libaccept.so

libaccept.so: accept.o
	$(LINK.c) -o $@ accept.o -ldl -lpthread
#-lasan

test: libaccept.so
	LD_PRELOAD=./$^ ls

libaccept.so *.o *.c: Makefile

clean:
	$(RM) libaccept.so accept.o
