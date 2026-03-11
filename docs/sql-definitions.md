# SQL Definitions

## Module Overview

This module does not directly use SQL databases for its core functionality.

## Database Usage

### Primary Storage
- **Technology**: In-memory registry
- **Purpose**: Tool handler registration and lookup
- **Schema**: No SQL schema is required

### Optional SQL Integration
ToolSchema could be extended to store tool usage statistics or tool configurations in a SQL database. If needed, use the `digital.vasic.database` module.

## Related Modules

For SQL database functionality, see the [Database module](../Database/README.md).

## Migration Notes

If SQL support is added in the future:
1. Create migration scripts in `migrations/` directory
2. Follow versioned migration pattern (`001_initial.sql`, `002_add_feature.sql`)
3. Use the `digital.vasic.database` module for database operations
4. Update this document with schema definitions