FROM python:3.12-slim
WORKDIR /app
RUN pip install --no-cache-dir psycopg2-binary
COPY scraper/seed.py .
COPY baserow_rows.json .
CMD ["python3", "seed.py"]
