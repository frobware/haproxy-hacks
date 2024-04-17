BUILD_DIR	= target/classes
CLASSPATH       = $(BUILD_DIR):$(shell cat classpath.txt 2>/dev/null)
JAVA		= java
JAVAC		= javac
MAIN_CLASS	= com.example.HttpClientSimulator
SOURCES         = $(shell find $(SRC_DIR) -type f -name '*.java')
SRC_DIR		= src/main/java

all: classpath.txt $(SOURCES)
	@mkdir -p $(BUILD_DIR)
	$(JAVAC) -cp "$(CLASSPATH)" -d $(BUILD_DIR) $(SOURCES)

run: all
	$(JAVA) -cp "$(CLASSPATH)" $(MAIN_CLASS) $(ROUTE)

clean:
	@rm -rf $(BUILD_DIR) classpath.txt

.PHONY: all run clean

# Target to generate classpath.txt using Maven.
classpath.txt:
	mvn dependency:build-classpath -Dmdep.outputFile=classpath.txt
