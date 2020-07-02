SQLITE3_OBJS := sqlite3.o shell.o
HISTO_OBJS := histogram.o helpers.o meanhistogram.o ratiohistogram.o SQLiteExt.o RegistExt.o

DEBUG ?= -O3

all: sqlite3 histograms.so

sqlite3: $(SQLITE3_OBJS) Makefile
	$(LINK.cpp) $(DEBUG) -o $@ $(SQLITE3_OBJS) -lm -lpthread -ldl

histograms.so: $(HISTO_OBJS)
	$(CXX) -fPIC $(DEBUG) -o $@ $(HISTO_OBJS) -lm -shared

%.o: %.c
	$(COMPILE.c) -o $@ $(CFLAGS) $(DEBUG) $<

%.o: %.cpp
	$(COMPILE.cpp) -shared -fPIC -o $@ $(DEBUG) $<

%.c: %.h

clean:
	$(RM) *.so *.o

$(SQLITE3_OBJS): Makefile
$(HISTO_OBJS): Makefile
