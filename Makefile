APP_NAME := falken
CORE_DIR := ../falken-core

.PHONY: all build core_assets clean run

all: core_assets build

core_assets:
	$(MAKE) -C $(CORE_DIR) default_tools default_plugins sync_embedded_assets plugins tools

build:
	go build -o $(APP_NAME) .

run: all
	./$(APP_NAME) --debug

clean:
	rm -f $(APP_NAME)
