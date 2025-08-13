#!/bin/bash

# Create test events for today
TODAY=$(date +%Y-%m-%d)
CALENDAR_DIR="$HOME/.bubblecal/days/$TODAY"

echo "Creating test events for $TODAY..."
mkdir -p "$CALENDAR_DIR"

# Morning meeting
cat > "$CALENDAR_DIR/0900-1000-Team_Standup" << EOF
category:work
description:Daily team sync
EOF

# Lunch
cat > "$CALENDAR_DIR/1200-1300-Lunch_Break" << EOF
category:personal
description:
EOF

# Afternoon meeting
cat > "$CALENDAR_DIR/1400-1500-Project_Review" << EOF
category:work
description:Review Q1 progress
EOF

# All-day event
cat > "$CALENDAR_DIR/allday-Company_Holiday" << EOF
category:holiday
description:Office closed
EOF

echo "Test events created!"
echo "Run ./bubblecal to test the calendar"