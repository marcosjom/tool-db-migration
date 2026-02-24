# tool-db-migration

Tool for migrating sql-server schema/data scripts to maria-db and other formats (if modified).

Created by [Marcos Ortega](https://mortegam.com/) for a specific migration of a system multi-GBs-databases from SQL Server to MariaDB. You can use it or modify it to your needs.

# Features

Input:

- Uses SqlServer Batch scripts as input:
   - Open you SQL Server Management Tool
   - Use the "Generate Script..." option
   - Edit the options to include Schema, Data, Foreign Keys and Indexes.
   - Keep the Unicode format.
   - Export.

Output:

- Applies fixes to known issues before parsing a batch.
- Supports:
  - Use Statements
  - Create Table Statements
  - Create Index Statements
  - Alter Table Statements
  - Insert Statements
- Converts types (ex: mssql-tinyint to mariadb-tinyint-unsigned).
- Creates unique indexes if and auto-increment field is not primery-key.
- Outputs MariaDB compatible sql.

# Compile it (Windows, Linux, Mac)

```
cd tool-db-migration
go get [github.com/marcosjom/tool-db-migration](https://github.com/marcosjom/tool-db-migration)
go build cmd/tsql2sql
```

# Run it (Windows, Linux, Mac)

```
./tsql2sql myFile.tsql > myNewFile.sql
```

# Dependencies

Thanks to [ha1tch](https://github.com/ha1tch) for his [tsql parser](https://github.com/ha1tch/tsqlparser).

# Contact

Visit [mortegam.com](https://mortegam.com/) to see this project running a website on Linux.

May you be surrounded by passionate and curious people. :-)



