APP_NAME="line-racer"

build: ## Build the container
	docker build -t $(APP_NAME) .

run: ## Run the container
	docker run -i -t --rm --name="$(APP_NAME)" $(APP_NAME)

up: build run

stop: ## stop and remove a running container
	docker stop $(APP_NAME); docker rm $(APP_NAME)