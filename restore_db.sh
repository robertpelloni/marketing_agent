#!/bin/bash
WSL_IP=$(hostname -I | awk '{print $1}')
export PGPASSWORD=secret

echo "Restoring database at $WSL_IP:5433..."
psql -h $WSL_IP -p 5433 -U torbnexus -d torbnexus -f /mnt/c/Users/hyper/workspace/enterprise_sales_bot/sales_bot_data.sql 2>&1 | tail -3
echo "=== Restore complete ==="

psql -h $WSL_IP -p 5433 -U torbnexus -d torbnexus -c "SELECT count(*) AS companies FROM companies;"
psql -h $WSL_IP -p 5433 -U torbnexus -d torbnexus -c "SELECT count(*) AS contacts FROM contacts;"
psql -h $WSL_IP -p 5433 -U torbnexus -d torbnexus -c "SELECT current_state, count(*) FROM deals GROUP BY current_state;"
