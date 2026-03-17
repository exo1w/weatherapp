from flask import Flask, jsonify, make_response, Response
from flask_cors import CORS
import requests
import os
import logging
from prometheus_client import Counter, Histogram, generate_latest
import time

app = Flask(__name__)
CORS(app)

logging.basicConfig(level=logging.INFO)

API_KEY = os.getenv("APIKEY")

# ===== Prometheus Metrics =====
REQUEST_COUNT = Counter(
    "weather_requests_total",
    "Total number of weather requests",
    ["city"]
)

REQUEST_LATENCY = Histogram(
    "weather_request_duration_seconds",
    "Weather request latency in seconds"
)
# ==============================

@app.route("/")
def health():
    return jsonify({
        "status": "ok",
        "service": "weather-service"
    }), 200


@app.errorhandler(Exception)
def handle_error(error):
    logging.exception("Unhandled exception occurred")
    response = {
        "message": "Internal server error",
        "error": str(error)
    }
    return make_response(jsonify(response), 500)


@app.route('/<city>')
def get_weather(city):
    start_time = time.time()
    REQUEST_COUNT.labels(city=city).inc()

    if not API_KEY:
        return make_response(jsonify({
            "message": "API key missing in environment variables"
        }), 500)

    url = "https://weatherapi-com.p.rapidapi.com/current.json"

    headers = {
        "x-rapidapi-host": "weatherapi-com.p.rapidapi.com",
        "x-rapidapi-key": API_KEY
    }

    params = {"q": city}

    try:
        response = requests.get(
            url,
            headers=headers,
            params=params,
            timeout=10
        )

        logging.info(f"API Status Code: {response.status_code}")
        logging.info(response.text)

        if response.status_code != 200:
            return make_response(jsonify({
                "message": "API request failed",
                "status": response.status_code,
                "response": response.text
            }), response.status_code)

        return jsonify(response.json())

    except requests.exceptions.Timeout:
        return make_response(jsonify({
            "message": "Request timeout"
        }), 504)

    except requests.exceptions.RequestException as e:
        return make_response(jsonify({
            "message": "Request error",
            "error": str(e)
        }), 500)

    finally:
        REQUEST_LATENCY.observe(time.time() - start_time)


# ===== Prometheus metrics endpoint =====
@app.route("/metrics")
def metrics():
    return Response(generate_latest(), mimetype="text/plain")
# ======================================

if __name__ == '__main__':
    app.run(host="0.0.0.0", port=5000)