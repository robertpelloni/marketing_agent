#!/bin/bash
# Apollo + Hunter Prospecting Script for TormentNexus
# Usage: ./scripts/apollo-hunter-prospect.sh

set -e

# Load environment
if [ -f .env ]; then
	source .env
fi

API_KEY="${APOLLO_API_KEY}"
HUNTER_KEY="${HUNTER_API_KEY}"

if [ -z "$API_KEY" ]; then
	echo "Error: APOLLO_API_KEY not set"
	exit 1
fi

if [ -z "$HUNTER_KEY" ]; then
	echo "Error: HUNTER_API_KEY not set"
	exit 1
fi

echo "=== Apollo + Hunter Prospecting for TormentNexus ==="
echo ""

# Function: Search Apollo for people
search_apollo() {
	local title="$1"
	local seniority="$2"
	local company_size="$3"
	local page="${4:-1}"

	echo "Searching Apollo for: $title ($seniority) at companies $company_size..."

	curl -s -X POST "https://api.apollo.io/api/v1/people/search" \
		-H "Content-Type: application/json" \
		-H "Cache-Control: no-cache" \
		-d '{
            "api_key": "'$API_KEY'",
            "q_keywords": "AI LLM machine learning",
            "title": "'$title'",
            "seniority": "'$seniority'",
            "organization_num_employees_ranges": ["'$company_size'"],
            "page": '$page',
            "per_page": 25
        }' | python3 -c "
import sys, json
try:
    data = json.load(sys.stdin)
    people = data.get('people', [])
    print(f'Found {len(people)} contacts')
    for p in people:
        name = p.get('name', 'Unknown')
        title = p.get('title', 'Unknown')
        org = p.get('organization', {}).get('name', 'Unknown')
        email = p.get('email', 'No email')
        linkedin = p.get('linkedin_url', '')
        print(f'  - {name} | {title} at {org} | {email}')
except Exception as e:
    print(f'Error: {e}')
"
}

# Function: Search Hunter for domain emails
search_hunter() {
	local domain="$1"

	echo "Searching Hunter for: $domain..."

	curl -s "https://api.hunter.io/v2/domain-search?domain=$domain&api_key=$HUNTER_KEY&limit=10&type=personal" | python3 -c "
import sys, json
try:
    data = json.load(sys.stdin)
    emails = data.get('data', {}).get('emails', [])
    print(f'Found {len(emails)} emails')
    for e in emails:
        email = e.get('value', 'Unknown')
        confidence = e.get('confidence', 0)
        position = e.get('position', 'Unknown')
        print(f'  - {email} (confidence: {confidence}%, position: {position})')
except Exception as e:
    print(f'Error: {e}')
"
}

# Function: Verify email with Hunter
verify_email() {
	local email="$1"

	echo "Verifying: $email..."

	curl -s "https://api.hunter.io/v2/email-verifier?email=$email&api_key=$HUNTER_KEY" | python3 -c "
import sys, json
try:
    data = json.load(sys.stdin)
    result = data.get('data', {})
    status = result.get('result', 'unknown')
    score = result.get('score', 0)
    print(f'Status: {status} | Score: {score}')
except Exception as e:
    print(f'Error: {e}')
"
}

echo "=== Step 1: Search Apollo for AI/ML Decision Makers ==="
echo ""

# Search for CTOs at AI startups
search_apollo "CTO" "c_suite" "11,50"

echo ""

# Search for VPs of Engineering
search_apollo "VP Engineering" "vp" "51,200"

echo ""

# Search for AI/ML Leads
search_apollo "AI Lead" "director" "51,200"

echo ""

echo "=== Step 2: Search Hunter for Target Domains ==="
echo ""

# Search common AI company domains
for domain in "cursor.sh" "continue.dev" "langchain.dev" "crewai.com" "openrouter.ai"; do
	search_hunter "$domain"
	echo ""
done

echo "=== Step 3: Verify Sample Emails ==="
echo ""

# Verify sample emails (replace with actual emails from Apollo)
verify_email "test@example.com"

echo ""
echo "=== Prospecting Complete ==="
echo ""
echo "Next steps:"
echo "1. Export results to CSV"
echo "2. Import to Apollo sequences"
echo "3. Launch outreach campaigns"
