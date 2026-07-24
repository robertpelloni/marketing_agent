#!/bin/bash
echo "=== NEW DEALS SAMPLE ==="
su - postgres -c "psql -d sales_bot -c \"SELECT d.id, c.name, d.current_state, d.cadence_step, co.email FROM deals d JOIN companies c ON d.company_id = c.id JOIN contacts co ON co.company_id = c.id WHERE d.id > 3450 ORDER BY d.id LIMIT 10;\""
echo "=== CONTACTS FOR NEW DEALS ==="
su - postgres -c "psql -d sales_bot -c \"SELECT COUNT(DISTINCT d.id) FROM deals d JOIN companies c ON d.company_id = c.id JOIN contacts co ON co.company_id = c.id WHERE d.id > 3450 AND d.cadence_step = 0 AND co.email != '';\""
