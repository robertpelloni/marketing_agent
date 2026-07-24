#!/bin/bash
# Force reset cadence for new deals and bump them to trigger immediate outreach
echo "=== Resetting cadence for new deals with contacts ==="
su - postgres -c "psql -d sales_bot -c \"UPDATE deals SET cadence_step = 0, current_state = 'Researched', updated_at = NOW() WHERE id > 3450 AND cadence_step = 0 AND current_state = 'Researched';\""
echo "=== Marking old exhausted deals as Outreach_Sent to skip them ==="
su - postgres -c "psql -d sales_bot -c \"UPDATE deals SET current_state = 'Outreach_Sent' WHERE cadence_step >= 3 AND current_state = 'Researched';\""
echo "=== New state distribution ==="
su - postgres -c "psql -d sales_bot -c \"SELECT current_state, cadence_step, COUNT(*) FROM deals GROUP BY current_state, cadence_step ORDER BY current_state, cadence_step;\""
