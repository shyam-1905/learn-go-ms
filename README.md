# 🚀 PROJECT: **Cloud-Native Expense Tracker Platform**

> **Goal:** Build a **production-grade, cloud-native, microservices-based Expense Tracker** using **Golang**, deployed on **AWS EKS**, with **Terraform**, **Docker**, **CI/CD**, and **AWS integrations**.

---

## 🧠 WHAT THIS PROJECT PROVES (Interview Mapping)

| Skill Required   | Covered By                                           |
| ---------------- | ---------------------------------------------------- |
| Golang expertise | Clean architecture, concurrency, context, interfaces |
| Microservices    | Multiple services, communication                     |
| AWS EKS          | Real deployment                                      |
| Docker           | Multi-stage builds                                   |
| Terraform        | Infra provisioning                                   |
| IAM & Security   | IRSA, least privilege                                |
| Python + Boto3   | Automation scripts                                   |
| CI/CD            | GitHub Actions                                       |
| Observability    | Logs, metrics, health checks                         |
| Secure cloud     | TLS, IAM, private networking                         |

---

# 🏗️ HIGH-LEVEL ARCHITECTURE

```
Client (Web / Curl / Postman)
        |
   API Gateway / ALB
        |
----------------------------
|        EKS Cluster        |
|                            |
|  ┌──────────────┐          |
|  │ Auth Service │          |
|  └──────────────┘          |
|          |                 |
|  ┌──────────────────┐      |
|  │ Expense Service  │───┐  |
|  └──────────────────┘   │  |
|          |              │  |
|  ┌──────────────────┐  │  |
|  │ Receipt Service  │◄─┘  |
|  └──────────────────┘      |
|                            |
----------------------------
     |        |         |
    RDS      S3      CloudWatch
```

---

# 🧩 MICROSERVICES BREAKDOWN

### **1️⃣ Auth Service**

**Purpose:** Authentication & authorization

* JWT generation
* Token validation
* User management
* Middleware for auth

**Tech**

* Golang
* JWT
* Postgres (users table)

---

### **2️⃣ Expense Service**

**Purpose:** Core business logic

* Create expense
* Get expenses
* Category filters
* Date range queries
* Pagination

**Tech**

* Golang
* Postgres
* Concurrency (batch operations)

---

### **3️⃣ Receipt Service**

**Purpose:** File uploads

* Upload receipt images
* Store in S3
* Generate pre-signed URLs
* Metadata storage

**Tech**

* Golang
* AWS SDK
* IRSA

---

### **4️⃣ Notification Service (Optional but Powerful)**

**Purpose:** Async processing

* Expense alerts
* Monthly summary
* Email triggers

**Tech**

* Golang
* Goroutines
* Channels
* SQS / SNS

---

# 📂 MONOREPO STRUCTURE (VERY IMPORTANT)

```
expense-tracker/
│
├── services/
│   ├── auth-service/
│   │   ├── cmd/
│   │   │   └── main.go
│   │   ├── internal/
│   │   │   ├── handler/
│   │   │   ├── service/
│   │   │   ├── repository/
│   │   │   ├── model/
│   │   │   └── middleware/
│   │   ├── Dockerfile
│   │   ├── go.mod
│   │   └── README.md
│   │
│   ├── expense-service/
│   │   ├── cmd/
│   │   │   └── main.go
│   │   ├── internal/
│   │   │   ├── handler/
│   │   │   ├── service/
│   │   │   ├── repository/
│   │   │   ├── model/
│   │   │   └── validator/
│   │   ├── Dockerfile
│   │   └── go.mod
│   │
│   ├── receipt-service/
│   │   ├── cmd/
│   │   │   └── main.go
│   │   ├── internal/
│   │   │   ├── handler/
│   │   │   ├── service/
│   │   │   ├── repository/
│   │   │   └── model/
│   │   ├── Dockerfile
│   │   └── go.mod
│   │
│   └── notification-service/
│       ├── cmd/
│       │   └── main.go
│       ├── internal/
│       │   ├── consumer/
│       │   ├── producer/
│       │   └── worker/
│       ├── Dockerfile
│       └── go.mod
│
├── deployments/
│   ├── auth-service/
│   │   ├── deployment.yaml
│   │   ├── service.yaml
│   │   ├── hpa.yaml
│   │   ├── configmap.yaml
│   │   ├── serviceaccount.yaml
│   │   └── ingress.yaml
│   │
│   ├── expense-service/
│   │   ├── deployment.yaml
│   │   ├── service.yaml
│   │   ├── hpa.yaml
│   │   ├── configmap.yaml
│   │   └── ingress.yaml
│   │
│   ├── receipt-service/
│   │   ├── deployment.yaml
│   │   ├── service.yaml
│   │   ├── hpa.yaml
│   │   ├── configmap.yaml
│   │   └── serviceaccount.yaml
│   │
│   └── ingress/
│       └── alb-ingress.yaml
│
├── infra/
│   └── terraform/
│       ├── vpc/
│       ├── eks/
│       ├── rds/
│       ├── s3/
│       ├── iam/
│       └── alb/
│
├── scripts/
│   └── boto3/
│
├── .github/
│   └── workflows/
│       ├── auth-ci.yaml
│       ├── expense-ci.yaml
│       └── receipt-ci.yaml
│
└── README.md
```

---

# 🧪 GOLANG CONCEPTS COVERED (EXPLICIT)

| Concept            | Where Used              |
| ------------------ | ----------------------- |
| Interfaces         | Repository pattern      |
| Structs & Methods  | Models, services        |
| Concurrency        | Notification service    |
| Channels           | Async processing        |
| Context            | Request lifecycle       |
| Error wrapping     | All services            |
| JSON & tags        | API layer               |
| Middlewares        | Auth, logging           |
| Testing            | Service & handler tests |
| Clean Architecture | internal/ separation    |

---

# 🔐 SECURITY DESIGN (INTERVIEW GOLD)

* IAM Roles for Service Accounts (IRSA)
* S3 access via IAM (no keys)
* JWT authentication
* Secrets in AWS Secrets Manager
* RDS in private subnet
* ALB TLS termination
* Network policies (optional)

---

# 🧰 TERRAFORM MODULES

```
infra/terraform/
├── modules/
│   ├── vpc/
│   ├── eks/
│   ├── iam/
│   ├── rds/
│   ├── s3/
│   ├── alb/
│   └── ecr/
```

---

# 🔁 CI/CD PIPELINE

### **GitHub Actions**

1. Run Go tests
2. Build Docker images
3. Push to ECR
4. Terraform plan/apply
5. Deploy to EKS
6. Run smoke tests

---

# 📊 OBSERVABILITY

* Health endpoints (`/health`)
* Readiness probes
* Structured logs
* CloudWatch
* Prometheus metrics (optional)
