# FlowForge Automation
<p>
FlowForge Automation is web platform for managing and executing automated workflows using Directed Acyclic Graph (DAG) definitions
</p>

## Technologies
- Backend: Rest API with Go (Fiber) + PostgreSQL + GORM
- Frontend: Next.js + Tailwind CSS
- Database: PostgreSQL 16

## Core Features
- Multi Tenancy: Isolated environments for different tenants.
- DAG Workflow: Create complex automation flows using JSON Directed Acyclic Graphs.
- Multiple Triggers: Execute workflows manually, via cron schedules, or through external Webhooks.
- Role Based Access Control: Granular permissions for Super Admin, Tenant Admin, Editor and Viewer.
- Realtime Monitoring: Track workflow executions live using SSE
- Automated Testing: Comprehensive unit tests and E2E API testing

## Prerequisites
<p>
Before beginning, ensure you have the following installed:
</p>

- [Go 1.23 or higher](https://go.dev/dl/)
- [Node.js 20 or higher](https://nodejs.org/en)
- [Docker Desktop](https://www.docker.com/products/docker-desktop/) (WSL integration for Windows)
- [Make](https://www.gnu.org/software/make/) (optional, for easier command execution)

## Run Applications
1. Clone the repository:

   ```bash
   $ git clone https://github.com/SutantoAdiNugroho/flowforge-automation.git
   $ cd flowforge-automation
   ```

2. Running with Makefile
    - Run unit tests
        ```bash
        $ make test
        ```

    - Run all tests including E2E API tests
        ```bash
        $ bash ./e2e.sh
        ```

    - Run all applications (frontend + backend + database)
        ```bash
        $ make run
        ```
        <p>
        After the application is running, you can access the frontend at `http://localhost:3000` and the backend API at `http://localhost:5000`.
        </p>

3. Running without Makefile
    <p>
    If make is not available, you can run it manually :
    </p>

    - Initiate dependencies
        ```bash
        $ cd backend && go mod tidy && go mod download
        $ cd ../frontend && npm install
        ```
    - Run backend unit tests
        ```bash
        $ cd backend && go test ./pkg/... -v
        ``` 
    - Run all applications using Docker Compose
        ```bash
        $ docker compose down --volumes
        $ docker compose up --build -d
        ``` 
        <p>
        Access the frontend at `http://localhost:3000` and the backend at `http://localhost:5000`.
        </p>

## Testing
1. Unit Test
    <p>Unit tests are written for the backend packages:</p>

    ```bash
    $ cd backend && go test ./pkg/... -v
    ```

2. API Test (E2E)
    <p>Full API integration tests are available in the backend/tests directory:</p>

    ```bash
    $ cd backend && go test ./tests/... -v
    ```

## Running with Docker Compose
```bash
$ docker compose up --build
```
<p> 
Available services:
</p>

- **PostgreSQL** -> `localhost:5432`
- **Backend** -> `http://localhost:5000`
- **Frontend** -> `http://localhost:3000`
