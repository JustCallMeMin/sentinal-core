import csv
import json
import uuid
import datetime
import os
import psycopg2
from psycopg2.extras import execute_batch

# Configuration
CSV_PATH = ".reports/ieee-fraud-detection/train_transaction.csv"
TENANT_ID = "00000000-0000-0000-0000-000000000001" 
START_DATE = datetime.datetime(2023, 1, 1)

def ingest_data():
    # Priority: Env variable > Docker default
    db_url = os.getenv("DATABASE_URL", "postgres://sentinal:password123@localhost:5432/sentinal_core?sslmode=disable")
    
    # If running inside Docker network, replace localhost with postgres
    if os.getenv("DOCKER_CONTAINER") == "true":
        db_url = db_url.replace("localhost", "postgres")

    print(f"-- Connecting to DB: {db_url.split('@')[-1]}") # Log without credentials
    conn = psycopg2.connect(db_url)
    cur = conn.cursor()
    
    # Pre-create Tenant
    cur.execute("INSERT INTO tenants (tenant_id, name, industry_segment) VALUES (%s, %s, %s) ON CONFLICT DO NOTHING", 
                (TENANT_ID, "IEEE Demo Tenant", "Fintech"))
                
    # CRITICAL: Create Partitions for this Tenant
    tables = ["transactions", "feature_snapshots", "decisions", "labels", "review_tasks", "notifications", "review_activity_logs"]
    for table in tables:
        cur.execute(f"CREATE TABLE IF NOT EXISTS {table}_{TENANT_ID.replace('-', '_')} PARTITION OF {table} FOR VALUES IN ('{TENANT_ID}')")

    conn.commit()

    print("-- Reading CSV & Preparing Batches...")
    
    batch_tx = []
    batch_features = []
    batch_labels = []
    
    BATCH_SIZE = 1000
    
    with open(CSV_PATH, 'r') as f:
        reader = csv.DictReader(f)
        count = 0
        
        for row in reader:
            transaction_id = str(uuid.uuid4())
            correlation_id = row['TransactionID']
            is_fraud = row['isFraud']
            
            # Calculate timestamp
            delta_seconds = int(float(row['TransactionDT']))
            occurred_at = START_DATE + datetime.timedelta(seconds=delta_seconds)
            
            # Extract features (V1-V339, C1-C14, D1-D15)
            features = {}
            for k, v in row.items():
                if k.startswith(('V', 'C', 'D', 'M', 'card', 'addr')):
                    features[k] = v
            
            # Prepare Transaction
            batch_tx.append((
                transaction_id, TENANT_ID, correlation_id, 
                float(row['TransactionAmt'] or 0), 'USD', occurred_at, "{}"
            ))
            
            # Prepare Features
            batch_features.append((
                transaction_id, TENANT_ID, "v1", json.dumps(features), "hash_placeholder"
            ))
            
            # Prepare Labels (only if Fraud)
            if is_fraud == '1':
                batch_labels.append((
                   str(uuid.uuid4()), transaction_id, TENANT_ID, 'manual', 'fraud' 
                ))
                
            count += 1
            if count % BATCH_SIZE == 0:
                # Flush Batch
                execute_batch(cur, "INSERT INTO transactions (transaction_id, tenant_id, correlation_id, amount, currency, occurred_at, payload) VALUES (%s, %s, %s, %s, %s, %s, %s) ON CONFLICT DO NOTHING", batch_tx)
                execute_batch(cur, "INSERT INTO feature_snapshots (transaction_id, tenant_id, schema_hash, features_json, feature_hash) VALUES (%s, %s, %s, %s, %s) ON CONFLICT DO NOTHING", batch_features)
                if batch_labels:
                    execute_batch(cur, "INSERT INTO labels (label_id, transaction_id, tenant_id, source, value) VALUES (%s, %s, %s, %s, %s) ON CONFLICT DO NOTHING", batch_labels)
                
                conn.commit()
                print(f"Ingested {count} rows...")
                batch_tx, batch_features, batch_labels = [], [], [] # Reset buffers
        
    cur.close()
    conn.close()
    print("Ingestion Complete!")

if __name__ == "__main__":
    ingest_data()
