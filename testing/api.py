from bottle import Bottle

app = Bottle()

@app.get("/api/v1")
def record():
    return {
          "name": "example.com",
          "type": "A",
          "value": "1.2.3.4"
            }


if __name__ == "__main__":
    from bottle import run
    run(app)
