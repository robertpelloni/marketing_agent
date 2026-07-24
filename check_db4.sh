#!/bin/bash
echo "=== CONTACTS WITH OUTBOUND ==="
su - postgres -c "psql -d sales_bot -c \"SELECT COUNT(DISTINCT contact_id) FROM interactions WHERE direction = 'Outbound';\""
echo "=== DEALS STEP 0 WITH INTERACTIONS ==="
su - postgres -c "psql -d sales_bot -c \"SELECT COUNT(DISTINCT d.id) FROM deals d JOIN companies c ON d.company_id = c.id JOIN contacts co ON co.company_id = c.id JOIN interactions i ON i.contact_id = co.id WHERE d.cadence_step = 0 AND i.direction = 'Outbound';\""
echo "=== DEALS STEP 0 WITHOUT INTERACTIONS ==="
su - postgres -c "psql -d sales_bot -c \"SELECT COUNT(DISTINCT d.id) FROM deals d WHERE d.cadence_step = 0 AND d.id NOT IN (SELECT DISTINCT d2.id FROM deals d2 JOIN companies c ON d2.company_id = c.id JOIN contacts co ON co.company_id = c.id JOIN interactions i ON i.contact_id = co.id WHERE i.direction = 'Outbound');\""
