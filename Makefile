run:
	DATABASE_URL="user=docker password=docker dbname=ris  host=localhost port=7070" go run src/api/main.go

db:
	docker build -t eg_postgresql build/db/
	docker run --publish 7070:5432 eg_postgresql
access_db:
	psql --port 7070 --host localhost --user docker ris

.PHONY: sim
sim:
	python3 sim/sim.py
