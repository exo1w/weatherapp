# Weather App – Microservices Project

A fully containerized **Weather Application** using a **Microservices architecture**, deployed on a **3-node Kubernetes cluster on AWS** with **Load Balancer**.

## 🌟 Project Overview

This project consists of the following services:

1. **Auth Service (Go)**

   * Handles user authentication and authorization.
   * Connects securely to MySQL database.
   * Provides JWT tokens for the UI service.

2. **Weather Service (Python)**

   * Fetches weather data from external APIs.
   * Provides weather information to the UI service.

3. **UI Service (Frontend)**

   * Interactive user interface.
   * Communicates with Auth Service for login/signup.
   * Fetches weather data from Weather Service.

4. **Database (MySQL)**

   * StatefulSet with persistent storage (AWS EBS).
   * Initialized via Kubernetes Job for DB and user creation.

---

## 🚀 Architecture & Kubernetes

* **Cluster:** AWS EC2 – 3 nodes (kubeadm)
* **Deployments:** Auth, Weather, UI (with replicas for scalability)
* **StatefulSet:** MySQL database
* **Services:** ClusterIP for internal communication
* **Ingress:** NGINX Ingress Controller to expose UI via HTTP
* **Load Balancer:** AWS ALB with Target Group to distribute traffic
* **Health Checks:** Liveness & Readiness probes for all services
* **Secrets:** Used for database passwords, JWT secret, and weather API key
* **Docker Images:** Built locally and pushed to private/public registry

### Kubernetes Files

All Kubernetes manifests are in the `k8s/` folder:

```
k8s/
├── mysql.yaml
├── auth.yaml
├── weather.yaml
└── ui.yaml
```

---

## 🔧 Getting Started

### Prerequisites

* Docker
* Kubernetes cluster (AWS or local)
* kubectl configured to your cluster
* Access to your Docker registry

### Deployment Steps

1. Apply MySQL StatefulSet & Job:

```bash
kubectl apply -f k8s/mysql.yaml
```

2. Deploy Auth Service:

```bash
kubectl apply -f k8s/auth.yaml
```

3. Deploy Weather Service:

```bash
kubectl apply -f k8s/weather.yaml
```

4. Deploy UI Service & Ingress:

```bash
kubectl apply -f k8s/ui.yaml
```

5. Access the application via AWS Load Balancer DNS:

```
http://myalb-1712604628.us-east-1.elb.amazonaws.com/
```

---

## 🛠️ Tech Stack

* **Languages:** Go, Python, JavaScript (Frontend)
* **Database:** MySQL
* **Containerization:** Docker
* **Orchestration:** Kubernetes (StatefulSet, Deployment, Service, Ingress)
* **Cloud:** AWS (EC2, EBS, ALB)
* **Secrets Management:** Kubernetes Secrets
* **Health Checks:** Liveness & Readiness probes
* **CI/CD Ready:** Docker images ready for registry deployment

---

## 📂 Project Structure

```
weatherapp-microservices/
├── auth-service/
│   └── Dockerfile
├── weather-service/
│   └── Dockerfile
├── ui-service/
│   └── Dockerfile
├── k8s/
│   ├── mysql.yaml
│   ├── auth.yaml
│   ├── weather.yaml
│   └── ui.yaml
├── README.md
```

---

## ✅ Features

* Microservices with independent scaling
* Secure communication with database using secrets
* High availability with multiple replicas
* Fully automated deployment on Kubernetes
* Publicly accessible via AWS Load Balancer
* Easy to extend with additional services or APIs
