# Use an official Python runtime as a parent image
FROM python:3.10-slim

# Set environment variables
ENV PYTHONDONTWRITEBYTECODE 1  # Prevent Python from writing .pyc files
ENV PYTHONUNBUFFERED 1        # Ensure stdout and stderr are flushed immediately

# Update package lists and install required dependencies
RUN apt-get update && \
    apt-get install -y --no-install-recommends \
    build-essential \
    iputils-ping \
    libpq-dev \
    curl \
    iproute2 \
    sqlite3 \
    && rm -rf /var/lib/apt/lists/*

# Set the working directory inside the container
WORKDIR /app

# Copy the requirements file and install dependencies
COPY requirements.txt .
RUN pip install --no-cache-dir -r requirements.txt

# Copy the rest of the application
COPY . .

# Expose the port your API will run on
EXPOSE 8081

# Command to run the application
# CMD uvicorn main:app --host 0.0.0.0 --port 8081
CMD ["ddtrace-run", "uvicorn", "main:app", "--host", "0.0.0.0", "--port", "8081", "--log-level", "info", "--reload"]
