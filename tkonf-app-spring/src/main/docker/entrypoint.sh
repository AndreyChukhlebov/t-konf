#!/bin/sh

# Default Java options
DEFAULT_JAVA_OPTS="-Dserver.address=0.0.0.0"

# Use JAVA_OPTS from environment if set, otherwise use defaults
JAVA_OPTS="${JAVA_OPTS:-$DEFAULT_JAVA_OPTS}"

# Add user-provided Java arguments
if [ -n "$JAVA_ARGS" ]; then
    JAVA_OPTS="$JAVA_OPTS $JAVA_ARGS"
fi

# Add additional Java options if needed
if [ -n "$ADDITIONAL_JAVA_OPTS" ]; then
    JAVA_OPTS="$JAVA_OPTS $ADDITIONAL_JAVA_OPTS"
fi

echo "Starting application with Java options: $JAVA_OPTS"

# Start the application
exec java $JAVA_OPTS -jar /deployments/tkonf-app-spring-0.0.1-SNAPSHOT.jar