SQLITE3_OBJS := sqlite3.o shell.o
DEBUG ?= -g
#CFLAGS ?= -DSQLITE_OMIT_LOAD_EXTENSION=0 -DSQLITE_THREADSAFE=0

sqlite3: $(SQLITE3_OBJS) Makefile
	$(LINK.cpp) $(DEBUG) -o $@ $(SQLITE3_OBJS) -lm -lpthread -ldl
	g++ -fPIC -lm -shared histogram.cpp helpers.cpp meanhistogram.cpp ratiohistogram.cpp SQLiteExt.cpp RegistExt.cpp -o histograms.so

$(SQLITE3_OBJS): Makefile

%.o: %.c
	$(COMPILE.c) -o $@ $(CFLAGS) $(DEBUG) $<

# %.o: %.cpp
# 	$(COMPILE.cpp) -fPIC -o $@ $(DEBUG) $<

%.c: %.h

clean:
	$(RM) *.so *.o
