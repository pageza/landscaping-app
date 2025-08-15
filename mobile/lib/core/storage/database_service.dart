import 'dart:async';
import 'dart:io';
import 'package:path/path.dart';
import 'package:sqflite/sqflite.dart';
import 'package:path_provider/path_provider.dart';
import 'package:logger/logger.dart';
import '../config/app_config.dart';
import 'database_schema.dart';

class DatabaseService {
  static DatabaseService? _instance;
  static Database? _database;
  final Logger _logger = Logger();

  DatabaseService._internal();

  factory DatabaseService() {
    _instance ??= DatabaseService._internal();
    return _instance!;
  }

  Future<Database> get database async {
    _database ??= await _initDatabase();
    return _database!;
  }

  Future<Database> _initDatabase() async {
    try {
      final documentsDirectory = await getApplicationDocumentsDirectory();
      final path = join(documentsDirectory.path, AppConfig.dbName);
      
      _logger.i('Initializing database at: $path');
      
      return await openDatabase(
        path,
        version: AppConfig.dbVersion,
        onCreate: DatabaseSchema.onCreate,
        onUpgrade: DatabaseSchema.onUpgrade,
        onConfigure: _onConfigure,
        onOpen: _onOpen,
      );
    } catch (e) {
      _logger.e('Error initializing database: $e');
      rethrow;
    }
  }

  Future<void> _onConfigure(Database db) async {
    // Enable foreign key constraints
    await db.execute('PRAGMA foreign_keys = ON');
    // Enable WAL mode for better performance
    await db.execute('PRAGMA journal_mode = WAL');
    // Set synchronous mode for better performance
    await db.execute('PRAGMA synchronous = NORMAL');
    // Set temp store to memory for better performance
    await db.execute('PRAGMA temp_store = MEMORY');
    // Set cache size (negative value means KB)
    await db.execute('PRAGMA cache_size = -64000'); // 64MB
  }

  Future<void> _onOpen(Database db) async {
    _logger.i('Database opened successfully');
    
    // Verify database integrity
    final result = await db.rawQuery('PRAGMA integrity_check');
    if (result.first['integrity_check'] != 'ok') {
      _logger.e('Database integrity check failed');
      throw Exception('Database integrity check failed');
    }
  }

  // Generic CRUD operations
  Future<List<Map<String, dynamic>>> query(
    String table, {
    bool? distinct,
    List<String>? columns,
    String? where,
    List<Object?>? whereArgs,
    String? groupBy,
    String? having,
    String? orderBy,
    int? limit,
    int? offset,
  }) async {
    final db = await database;
    return await db.query(
      table,
      distinct: distinct,
      columns: columns,
      where: where,
      whereArgs: whereArgs,
      groupBy: groupBy,
      having: having,
      orderBy: orderBy,
      limit: limit,
      offset: offset,
    );
  }

  Future<List<Map<String, dynamic>>> rawQuery(
    String sql, [
    List<Object?>? arguments,
  ]) async {
    final db = await database;
    return await db.rawQuery(sql, arguments);
  }

  Future<int> insert(
    String table,
    Map<String, Object?> values, {
    String? nullColumnHack,
    ConflictAlgorithm? conflictAlgorithm,
  }) async {
    final db = await database;
    return await db.insert(
      table,
      values,
      nullColumnHack: nullColumnHack,
      conflictAlgorithm: conflictAlgorithm ?? ConflictAlgorithm.replace,
    );
  }

  Future<int> update(
    String table,
    Map<String, Object?> values, {
    String? where,
    List<Object?>? whereArgs,
    ConflictAlgorithm? conflictAlgorithm,
  }) async {
    final db = await database;
    return await db.update(
      table,
      values,
      where: where,
      whereArgs: whereArgs,
      conflictAlgorithm: conflictAlgorithm,
    );
  }

  Future<int> delete(
    String table, {
    String? where,
    List<Object?>? whereArgs,
  }) async {
    final db = await database;
    return await db.delete(
      table,
      where: where,
      whereArgs: whereArgs,
    );
  }

  Future<void> execute(String sql, [List<Object?>? arguments]) async {
    final db = await database;
    await db.execute(sql, arguments);
  }

  // Batch operations
  Future<List<Object?>> batch(Function(Batch batch) operations) async {
    final db = await database;
    final batch = db.batch();
    operations(batch);
    return await batch.commit();
  }

  // Transaction support
  Future<T> transaction<T>(Future<T> Function(Transaction txn) action) async {
    final db = await database;
    return await db.transaction(action);
  }

  // Utility methods
  Future<bool> tableExists(String tableName) async {
    try {
      final result = await rawQuery(
        "SELECT name FROM sqlite_master WHERE type='table' AND name=?",
        [tableName],
      );
      return result.isNotEmpty;
    } catch (e) {
      _logger.e('Error checking if table exists: $e');
      return false;
    }
  }

  Future<List<String>> getTableNames() async {
    try {
      final result = await rawQuery(
        "SELECT name FROM sqlite_master WHERE type='table' AND name NOT LIKE 'sqlite_%'",
      );
      return result.map((row) => row['name'] as String).toList();
    } catch (e) {
      _logger.e('Error getting table names: $e');
      return [];
    }
  }

  Future<int> getTableRowCount(String tableName) async {
    try {
      final result = await rawQuery('SELECT COUNT(*) as count FROM $tableName');
      return result.first['count'] as int;
    } catch (e) {
      _logger.e('Error getting row count for table $tableName: $e');
      return 0;
    }
  }

  Future<Map<String, int>> getAllTableRowCounts() async {
    final tableNames = await getTableNames();
    final counts = <String, int>{};
    
    for (final tableName in tableNames) {
      counts[tableName] = await getTableRowCount(tableName);
    }
    
    return counts;
  }

  // Database maintenance
  Future<void> vacuum() async {
    try {
      _logger.i('Running database vacuum...');
      await execute('VACUUM');
      _logger.i('Database vacuum completed');
    } catch (e) {
      _logger.e('Error running vacuum: $e');
    }
  }

  Future<void> analyze() async {
    try {
      _logger.i('Running database analyze...');
      await execute('ANALYZE');
      _logger.i('Database analyze completed');
    } catch (e) {
      _logger.e('Error running analyze: $e');
    }
  }

  Future<int> getDatabaseSize() async {
    try {
      final documentsDirectory = await getApplicationDocumentsDirectory();
      final path = join(documentsDirectory.path, AppConfig.dbName);
      final file = File(path);
      
      if (await file.exists()) {
        return await file.length();
      }
      return 0;
    } catch (e) {
      _logger.e('Error getting database size: $e');
      return 0;
    }
  }

  String formatDatabaseSize(int sizeInBytes) {
    const int kb = 1024;
    const int mb = kb * 1024;
    const int gb = mb * 1024;

    if (sizeInBytes >= gb) {
      return '${(sizeInBytes / gb).toStringAsFixed(2)} GB';
    } else if (sizeInBytes >= mb) {
      return '${(sizeInBytes / mb).toStringAsFixed(2)} MB';
    } else if (sizeInBytes >= kb) {
      return '${(sizeInBytes / kb).toStringAsFixed(2)} KB';
    } else {
      return '$sizeInBytes bytes';
    }
  }

  // Database backup and restore
  Future<String> backup() async {
    try {
      final documentsDirectory = await getApplicationDocumentsDirectory();
      final sourcePath = join(documentsDirectory.path, AppConfig.dbName);
      final backupPath = join(
        documentsDirectory.path,
        'backups',
        '${AppConfig.dbName}.backup.${DateTime.now().millisecondsSinceEpoch}',
      );

      final backupDir = Directory(dirname(backupPath));
      if (!await backupDir.exists()) {
        await backupDir.create(recursive: true);
      }

      final sourceFile = File(sourcePath);
      if (await sourceFile.exists()) {
        await sourceFile.copy(backupPath);
        _logger.i('Database backed up to: $backupPath');
        return backupPath;
      } else {
        throw Exception('Source database file does not exist');
      }
    } catch (e) {
      _logger.e('Error creating database backup: $e');
      rethrow;
    }
  }

  Future<void> restore(String backupPath) async {
    try {
      await close();
      
      final documentsDirectory = await getApplicationDocumentsDirectory();
      final targetPath = join(documentsDirectory.path, AppConfig.dbName);
      
      final backupFile = File(backupPath);
      if (await backupFile.exists()) {
        await backupFile.copy(targetPath);
        _logger.i('Database restored from: $backupPath');
        
        // Reinitialize the database
        _database = null;
        await database;
      } else {
        throw Exception('Backup file does not exist');
      }
    } catch (e) {
      _logger.e('Error restoring database: $e');
      rethrow;
    }
  }

  // Database reset
  Future<void> reset() async {
    try {
      final db = await database;
      await DatabaseSchema.resetDatabase(db);
      _logger.i('Database reset completed');
    } catch (e) {
      _logger.e('Error resetting database: $e');
      rethrow;
    }
  }

  Future<void> close() async {
    if (_database != null) {
      await _database!.close();
      _database = null;
      _logger.i('Database closed');
    }
  }

  // Helper method for debugging
  Future<void> printDatabaseInfo() async {
    try {
      final size = await getDatabaseSize();
      final tableNames = await getTableNames();
      final counts = await getAllTableRowCounts();
      
      _logger.i('=== Database Info ===');
      _logger.i('Size: ${formatDatabaseSize(size)}');
      _logger.i('Tables: ${tableNames.length}');
      
      for (final tableName in tableNames) {
        _logger.i('  $tableName: ${counts[tableName]} rows');
      }
      _logger.i('====================');
    } catch (e) {
      _logger.e('Error printing database info: $e');
    }
  }
}