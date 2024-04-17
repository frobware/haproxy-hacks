JAVAC = javac
JAVA = java
SRC_DIR = src/main/java
BUILD_DIR= target/classes
MAIN_CLASS = com.example.Main

CLASSPATH=$(BUILD_DIR):lib/*

SOURCES=$(shell find $(SRC_DIR) -name '*.java')

# Targets
all: $(SOURCES)
	@mkdir -p $(BUILD_DIR)
	$(JAVAC) -cp $(CLASSPATH) -d $(BUILD_DIR) $(SOURCES)

run: all
	$(JAVA) -cp $(CLASSPATH) $(MAIN_CLASS)

clean:
	@rm -rf $(BUILD_DIR)

.PHONY: all run clean

