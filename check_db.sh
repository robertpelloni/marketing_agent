#!/bin/bash
su - postgres -c "psql -d sales_bot -c '\\dt'"
echo "---"
su - postgres -c "psql -d sales_bot -c 'SELECT COUNT(*) FROM contacts;'"
echo "---"
su - postgres -c "psql -d sales_bot -c 'SELECT COUNT(*) FROM deals;'"
