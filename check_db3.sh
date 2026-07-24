#!/bin/bash
echo "=== INTERACTIONS SAMPLE ==="
su - postgres -c "psql -d sales_bot -c \"SELECT id, contact_id, channel, direction, summary, created_at FROM interactions ORDER BY created_at DESC LIMIT 10;\""
echo "=== OUTBOUND COUNT ==="
su - postgres -c "psql -d sales_bot -c \"SELECT COUNT(*) FROM interactions WHERE direction ILIKE '%outbound%';\""
echo "=== CADENCE STEP 0 DEALS WITH CONTACTS ==="
su - postgres -c "psql -d sales_bot -c \"SELECT d.id, c.name, co.email, co.id as contact_id FROM deals d JOIN companies c ON d.company_id = c.id JOIN contacts co ON co.company_id = c.id WHERE d.cadence_step = 0 AND d.current_state = 'Researched' LIMIT 5;\""
echo "=== INTERACTIONS FOR CONTACT 1 ==="
su - postgres -c "psql -d sales_bot -c \"SELECT * FROM interactions WHERE contact_id = 1;\""
