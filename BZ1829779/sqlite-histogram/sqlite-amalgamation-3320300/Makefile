SQLITE3_OBJS := sqlite3.o shell.o
HISTO_OBJS   := histogram.o helpers.o meanhistogram.o ratiohistogram.o SQLiteExt.o RegistExt.o
EXT_OBJS     := extension-functions.o

DEBUG ?= -O3

all: sqlite3

sqlite3: $(SQLITE3_OBJS) $(HISTO_OBJS) $(EXT_OBJS) Makefile
	$(LINK.cpp) $(DEBUG) -o $@ $(SQLITE3_OBJS) $(HISTO_OBJS) $(EXT_OBJS) -lm -lpthread -ldl

# histograms.so: $(HISTO_OBJS)
# 	$(CXX) $(DEBUG) -o $@ $(HISTO_OBJS) -lm -shared

# sqlitefunctions.so: $(EXT_OBJS)
# 	$(CXX) $(DEBUG) -o $@ $(EXT_OBJS) -lm -shared

%.o: %.c
	$(COMPILE.c) -o $@ $(CFLAGS) $(DEBUG) $<

%.o: %.cpp
	$(COMPILE.cpp) -o $@ $(DEBUG) -DSQLITE_CORE $<

%.c: %.h

clean:
	$(RM) *.so *.o

$(SQLITE3_OBJS): Makefile
$(HISTO_OBJS): Makefile
$(EXT_OBJS): Makefile
