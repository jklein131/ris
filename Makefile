run:
	DATABASE_URL="user=docker password=docker dbname=ris  host=localhost port=7070" go run src/api/main.go

run.docker: stop.all db.docker
	docker build -t main_api .
	docker run --publish 8080:8080 --link db -it --name api --rm -e DATABASE_URL="user=docker password=docker dbname=ris  host=db port=5432"  main_api

run.daemon.docker: stop.all db.docker
	docker build -t main_api .
	docker run --publish 8080:8080 --link db -d -it --name api --rm -e DATABASE_URL="user=docker password=docker dbname=ris  host=db port=5432"  main_api

stop.all:
	docker stop api || true
	docker stop db || true
	docker rm api || true
	docker rm db || true

db.docker:
	docker build -t main_pg build/db/
	docker run --name db -d --publish 7070:5432 main_pg || true

access_db:
	psql --port 7070 --host localhost --user docker ris

.PHONY: sim.docker
sim.docker: run.daemon.docker
	python3 sim/sim.py
