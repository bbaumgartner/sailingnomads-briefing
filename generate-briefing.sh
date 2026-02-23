#!/bin/bash

# generate-briefing.sh
# Generates a daily briefing for the sailing logbook.
# Reads context from Logseq journal files, calls the Go briefing generator,
# and writes the output into today's journal file.
# Cross-platform: supports both macOS (Darwin) and Linux.

set -e

SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"

GREEN='\033[0;32m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
NC='\033[0m'

# --- Usage ---

if [ "$#" -lt 1 ] || [ "$#" -gt 2 ]; then
    echo -e "${RED}Usage: $0 <saillog_directory> [config_file]${NC}"
    echo "Example: $0 /Users/benno/Documents/saillog ./config.env"
    echo ""
    echo "The saillog_directory must contain a journals/ subdirectory."
    echo "The optional config_file sets LANG, CONTEXT_DAYS, BRIEFING_LOOKBACK_DAYS."
    exit 1
fi

SAILLOG_DIR="$1"
CONFIG_FILE="${2:-}"
JOURNALS_DIR="$SAILLOG_DIR/journals"

if [ ! -d "$JOURNALS_DIR" ]; then
    echo -e "${RED}Error: journals directory not found at $JOURNALS_DIR${NC}"
    exit 1
fi

# --- Defaults ---

LANG="de"
CONTEXT_DAYS=7
BRIEFING_LOOKBACK_DAYS=3

if [ -n "$CONFIG_FILE" ] && [ -f "$CONFIG_FILE" ]; then
    echo -e "${GREEN}Loading config from $CONFIG_FILE${NC}"
    source "$CONFIG_FILE"
fi

# --- Date helpers ---

today_date() {
    date +%Y-%m-%d
}

today_file_date() {
    date +%Y_%m_%d
}

# Get journal filename for N days ago
journal_file_for_days_ago() {
    local days_ago=$1
    local OS_TYPE
    OS_TYPE=$(uname)
    if [[ "$OS_TYPE" == "Darwin" ]]; then
        date -v-"${days_ago}"d +%Y_%m_%d
    else
        date -d "$days_ago days ago" +%Y_%m_%d
    fi
}

# --- Find GPS position from recent journal entries ---

find_gps_position() {
    echo -e "${YELLOW}Searching for current_position in recent journal entries...${NC}"

    for days_ago in $(seq 0 30); do
        local file_date
        file_date=$(journal_file_for_days_ago "$days_ago")
        local journal_file="$JOURNALS_DIR/${file_date}.md"

        if [ -f "$journal_file" ]; then
            local position
            position=$(grep -o 'current_position:: [0-9.-]*/[0-9.-]*' "$journal_file" 2>/dev/null | head -1 | sed 's/current_position:: //')

            if [ -n "$position" ]; then
                LATITUDE=$(echo "$position" | cut -d'/' -f1)
                LONGITUDE=$(echo "$position" | cut -d'/' -f2)
                echo -e "${GREEN}Found position: $LATITUDE, $LONGITUDE (from ${file_date}.md)${NC}"
                return 0
            fi
        fi
    done

    echo -e "${RED}Error: No current_position:: found in any journal entry from the last 30 days${NC}"
    echo "Please add a line like this to a recent journal entry:"
    echo "  - current_position:: 47.13826/8.60032"
    return 1
}

# --- Extract previous briefings ---

extract_previous_briefings() {
    local context=""

    for days_ago in $(seq 1 "$BRIEFING_LOOKBACK_DAYS"); do
        local file_date
        file_date=$(journal_file_for_days_ago "$days_ago")
        local journal_file="$JOURNALS_DIR/${file_date}.md"
        local display_date
        display_date=$(echo "$file_date" | tr '_' '-')

        if [ -f "$journal_file" ]; then
            local briefing
            briefing=$(extract_briefing_block "$journal_file")
            if [ -n "$briefing" ]; then
                context+="--- ${display_date} ---"$'\n'
                context+="$briefing"$'\n\n'
            fi
        fi
    done

    echo "$context"
}

# Extract the [[Tagesbriefing]] block and all its indented children from a file.
extract_briefing_block() {
    local file="$1"
    local in_block=false
    local result=""

    while IFS= read -r line; do
        if [[ "$line" =~ ^-\ \[\[Tagesbriefing\]\] ]]; then
            in_block=true
            result+="$line"$'\n'
            continue
        fi

        if [ "$in_block" = true ]; then
            # Block continues as long as lines are indented (start with tab or spaces followed by -)
            if [[ "$line" =~ ^[[:space:]]+(- |\t) ]] || [[ "$line" =~ ^$'\t' ]]; then
                result+="$line"$'\n'
            else
                # Non-indented line means the block ended
                break
            fi
        fi
    done < "$file"

    echo "$result"
}

# --- Extract recent logbook entries (excluding briefing blocks) ---

extract_logbook_context() {
    local context=""

    for days_ago in $(seq 0 "$CONTEXT_DAYS"); do
        local file_date
        file_date=$(journal_file_for_days_ago "$days_ago")
        local journal_file="$JOURNALS_DIR/${file_date}.md"
        local display_date
        display_date=$(echo "$file_date" | tr '_' '-')

        if [ -f "$journal_file" ]; then
            local content
            content=$(extract_non_briefing_content "$journal_file")
            if [ -n "$content" ]; then
                context+="--- ${display_date} ---"$'\n'
                context+="$content"$'\n\n'
            fi
        fi
    done

    echo "$context"
}

# Extract all content from a journal file EXCEPT the [[Tagesbriefing]] block.
extract_non_briefing_content() {
    local file="$1"
    local in_briefing=false
    local result=""

    while IFS= read -r line; do
        if [[ "$line" =~ ^-\ \[\[Tagesbriefing\]\] ]]; then
            in_briefing=true
            continue
        fi

        if [ "$in_briefing" = true ]; then
            if [[ "$line" =~ ^[[:space:]]+(- |\t) ]] || [[ "$line" =~ ^$'\t' ]]; then
                continue
            else
                in_briefing=false
            fi
        fi

        if [ "$in_briefing" = false ]; then
            result+="$line"$'\n'
        fi
    done < "$file"

    # Trim trailing whitespace
    echo "$result" | sed -e 's/[[:space:]]*$//'
}

# --- Build context for stdin ---

build_context() {
    local context=""

    local briefings
    briefings=$(extract_previous_briefings)
    if [ -n "$briefings" ]; then
        context+="=== PREVIOUS BRIEFINGS ==="$'\n'
        context+="$briefings"$'\n'
    fi

    local logbook
    logbook=$(extract_logbook_context)
    if [ -n "$logbook" ]; then
        context+="=== LOGBOOK CONTEXT ==="$'\n'
        context+="$logbook"
    fi

    echo "$context"
}

# --- Write briefing to journal ---

write_to_journal() {
    local briefing="$1"
    local today
    today=$(today_file_date)
    local journal_file="$JOURNALS_DIR/${today}.md"

    if [ -f "$journal_file" ]; then
        # Check if a briefing block already exists
        if grep -q '^\- \[\[Tagesbriefing\]\]' "$journal_file"; then
            echo -e "${YELLOW}Replacing existing briefing in ${today}.md${NC}"
            replace_briefing_in_file "$journal_file" "$briefing"
        else
            echo -e "${GREEN}Appending briefing to ${today}.md${NC}"
            echo "" >> "$journal_file"
            echo "$briefing" >> "$journal_file"
        fi
    else
        echo -e "${GREEN}Creating new journal file ${today}.md${NC}"
        echo "$briefing" > "$journal_file"
    fi
}

# Replace an existing [[Tagesbriefing]] block in a journal file.
replace_briefing_in_file() {
    local file="$1"
    local new_briefing="$2"
    local temp_file
    temp_file=$(mktemp)
    local in_briefing=false
    local replaced=false

    while IFS= read -r line; do
        if [[ "$line" =~ ^-\ \[\[Tagesbriefing\]\] ]]; then
            in_briefing=true
            if [ "$replaced" = false ]; then
                echo "$new_briefing" >> "$temp_file"
                replaced=true
            fi
            continue
        fi

        if [ "$in_briefing" = true ]; then
            if [[ "$line" =~ ^[[:space:]]+(- |\t) ]] || [[ "$line" =~ ^$'\t' ]]; then
                continue
            else
                in_briefing=false
            fi
        fi

        if [ "$in_briefing" = false ]; then
            echo "$line" >> "$temp_file"
        fi
    done < "$file"

    mv "$temp_file" "$file"
}

# --- Main ---

echo -e "${GREEN}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
echo -e "${GREEN}  Sailing Nomads Daily Briefing Generator${NC}"
echo -e "${GREEN}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
echo -e "Date: $(today_date)"
echo -e "Saillog: ${YELLOW}$SAILLOG_DIR${NC}"
echo -e "Language: ${YELLOW}$LANG${NC}"
echo ""

# Step 1: Find GPS position
find_gps_position
echo ""

# Step 2: Build context from logbook
echo -e "${YELLOW}Building context from logbook...${NC}"
CONTEXT=$(build_context)
CONTEXT_LINES=$(echo "$CONTEXT" | wc -l | tr -d ' ')
echo -e "${GREEN}Context: ${CONTEXT_LINES} lines${NC}"
echo ""

# Step 3: Call the Go program
echo -e "${YELLOW}Generating briefing...${CONTEXT} ${NC}"
BRIEFING=$(echo "$CONTEXT" | (cd "$SCRIPT_DIR" && go run . --lat "$LATITUDE" --lon "$LONGITUDE" --lang "$LANG" --prompt "$SCRIPT_DIR/prompt.md"))

if [ -z "$BRIEFING" ]; then
    echo -e "${RED}Error: briefing generation returned empty output${NC}"
    exit 1
fi

BRIEFING_LINES=$(echo "$BRIEFING" | wc -l | tr -d ' ')
echo -e "${GREEN}Briefing generated: ${BRIEFING} lines${NC}"
echo ""

# Step 4: Write to journal
write_to_journal "$BRIEFING"

echo ""
echo -e "${GREEN}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
echo -e "${GREEN}  Briefing complete!${NC}"
echo -e "${GREEN}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
