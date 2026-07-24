#!/bin/bash
echo "=== DEALS BY STATE ==="
su - postgres -c "psql -d sales_bot -c \"SELECT current_state, COUNT(*) FROM deals GROUP BY current_state;\""
echo "=== DEALS RESERCHED STEP 0 ==="
su - postgres -c "psql -d sales_bot -c \"SELECT COUNT(*) FROM deals WHERE current_state = 'Researched' AND cadence_step = 0;\""
echo "=== CONTACTS FOR RESERCHED STEP 0 ==="
su - postgres -c "psql -d sales_bot -c \"SELECT COUNT(DISTINCT co.id) FROM deals d JOIN companies c ON d.company_id = c.id JOIN contacts co ON co.company_id = c.id WHERE d.current_state = 'Researched' AND d.cadence_step = 0;\""
echo "=== CONTACTS WITH EMAILS STEP 0 ==="
su - postgres -c "psql -d sales_bot -c \"SELECT COUNT(DISTINCT co.id) FROM deals d JOIN companies c ON d.company_id = c.id JOIN contacts co ON co.company_id = c.id WHERE d.current_state = 'Researched' AND d.cadence_step = 0 AND co.email != '';\""
