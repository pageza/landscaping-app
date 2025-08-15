import 'package:sqflite/sqflite.dart';

class DatabaseSchema {
  static const String databaseName = 'landscaping_app.db';
  static const int databaseVersion = 1;

  static Future<void> onCreate(Database db, int version) async {
    await _createTables(db);
    await _createIndexes(db);
  }

  static Future<void> onUpgrade(Database db, int oldVersion, int newVersion) async {
    // Handle database upgrades here
    if (oldVersion < newVersion) {
      // Add migration logic as needed
    }
  }

  static Future<void> _createTables(Database db) async {
    // Jobs table
    await db.execute('''
      CREATE TABLE jobs (
        id TEXT PRIMARY KEY,
        customer_id TEXT NOT NULL,
        customer_name TEXT NOT NULL,
        customer_email TEXT NOT NULL,
        customer_phone TEXT,
        service_id TEXT NOT NULL,
        service_name TEXT NOT NULL,
        service_description TEXT,
        status TEXT NOT NULL,
        job_type TEXT NOT NULL,
        priority TEXT NOT NULL,
        scheduled_date TEXT NOT NULL,
        started_at TEXT,
        completed_at TEXT,
        cancelled_at TEXT,
        location_data TEXT NOT NULL,
        notes TEXT,
        customer_notes TEXT,
        crew_notes TEXT,
        assigned_crew_ids TEXT,
        pricing_data TEXT,
        estimated_duration REAL,
        actual_duration REAL,
        equipment_needed TEXT,
        tags TEXT,
        metadata TEXT,
        requires_signature INTEGER DEFAULT 0,
        customer_signature TEXT,
        crew_signature TEXT,
        created_at TEXT,
        updated_at TEXT,
        company_id TEXT,
        is_synced INTEGER DEFAULT 0,
        last_sync_at TEXT,
        local_id TEXT,
        needs_upload INTEGER DEFAULT 0
      )
    ''');

    // Job photos table
    await db.execute('''
      CREATE TABLE job_photos (
        id TEXT PRIMARY KEY,
        job_id TEXT NOT NULL,
        url TEXT,
        local_path TEXT,
        type TEXT NOT NULL,
        caption TEXT,
        description TEXT,
        taken_at TEXT,
        taken_by TEXT,
        latitude REAL,
        longitude REAL,
        is_uploaded INTEGER DEFAULT 0,
        uploaded_at TEXT,
        created_at TEXT,
        needs_upload INTEGER DEFAULT 1,
        FOREIGN KEY (job_id) REFERENCES jobs (id) ON DELETE CASCADE
      )
    ''');

    // Job time entries table
    await db.execute('''
      CREATE TABLE job_time_entries (
        id TEXT PRIMARY KEY,
        job_id TEXT NOT NULL,
        crew_member_id TEXT NOT NULL,
        crew_member_name TEXT NOT NULL,
        start_time TEXT NOT NULL,
        end_time TEXT,
        break_duration REAL,
        notes TEXT,
        latitude REAL,
        longitude REAL,
        is_manual_entry INTEGER DEFAULT 0,
        created_at TEXT,
        updated_at TEXT,
        is_synced INTEGER DEFAULT 0,
        local_id TEXT,
        FOREIGN KEY (job_id) REFERENCES jobs (id) ON DELETE CASCADE
      )
    ''');

    // Services table
    await db.execute('''
      CREATE TABLE services (
        id TEXT PRIMARY KEY,
        name TEXT NOT NULL,
        description TEXT NOT NULL,
        category_data TEXT,
        job_type TEXT NOT NULL,
        base_price REAL NOT NULL,
        price_unit TEXT,
        estimated_duration REAL,
        duration_unit TEXT,
        required_equipment TEXT,
        optional_equipment TEXT,
        skills_required TEXT,
        seasonality TEXT,
        is_active INTEGER DEFAULT 1,
        image_url TEXT,
        tags TEXT,
        metadata TEXT,
        company_id TEXT,
        created_at TEXT,
        updated_at TEXT,
        is_synced INTEGER DEFAULT 0,
        last_sync_at TEXT
      )
    ''');

    // Equipment table
    await db.execute('''
      CREATE TABLE equipment (
        id TEXT PRIMARY KEY,
        name TEXT NOT NULL,
        model TEXT NOT NULL,
        brand TEXT,
        serial_number TEXT,
        type TEXT NOT NULL,
        status TEXT NOT NULL,
        description TEXT,
        purchase_date TEXT,
        purchase_price REAL,
        supplier TEXT,
        warranty_info TEXT,
        warranty_expiry TEXT,
        location TEXT,
        assigned_to TEXT,
        last_maintenance_date TEXT,
        next_maintenance_date TEXT,
        qr_code TEXT,
        image_url TEXT,
        tags TEXT,
        specifications TEXT,
        metadata TEXT,
        company_id TEXT,
        is_active INTEGER DEFAULT 1,
        created_at TEXT,
        updated_at TEXT,
        is_synced INTEGER DEFAULT 0,
        last_sync_at TEXT
      )
    ''');

    // Equipment maintenance records table
    await db.execute('''
      CREATE TABLE maintenance_records (
        id TEXT PRIMARY KEY,
        equipment_id TEXT NOT NULL,
        type TEXT NOT NULL,
        date TEXT NOT NULL,
        description TEXT NOT NULL,
        performed_by TEXT,
        cost REAL,
        notes TEXT,
        parts_replaced TEXT,
        next_service_date TEXT,
        invoice_number TEXT,
        vendor_info TEXT,
        created_at TEXT,
        is_synced INTEGER DEFAULT 0,
        FOREIGN KEY (equipment_id) REFERENCES equipment (id) ON DELETE CASCADE
      )
    ''');

    // Equipment usage table
    await db.execute('''
      CREATE TABLE equipment_usage (
        id TEXT PRIMARY KEY,
        equipment_id TEXT NOT NULL,
        job_id TEXT NOT NULL,
        used_by TEXT NOT NULL,
        start_time TEXT NOT NULL,
        end_time TEXT,
        hours_used REAL,
        fuel_used REAL,
        condition TEXT,
        notes TEXT,
        issues_reported TEXT,
        created_at TEXT,
        is_synced INTEGER DEFAULT 0,
        FOREIGN KEY (equipment_id) REFERENCES equipment (id) ON DELETE CASCADE,
        FOREIGN KEY (job_id) REFERENCES jobs (id) ON DELETE CASCADE
      )
    ''');

    // Notifications table
    await db.execute('''
      CREATE TABLE notifications (
        id TEXT PRIMARY KEY,
        title TEXT NOT NULL,
        message TEXT NOT NULL,
        type TEXT NOT NULL,
        priority TEXT NOT NULL,
        user_id TEXT NOT NULL,
        job_id TEXT,
        related_entity_id TEXT,
        related_entity_type TEXT,
        data TEXT,
        image_url TEXT,
        action_url TEXT,
        actions TEXT,
        is_read INTEGER DEFAULT 0,
        is_archived INTEGER DEFAULT 0,
        read_at TEXT,
        scheduled_for TEXT,
        expires_at TEXT,
        channel TEXT,
        is_delivered INTEGER DEFAULT 1,
        delivered_at TEXT,
        device_token TEXT,
        platform_type TEXT,
        created_at TEXT,
        updated_at TEXT
      )
    ''');

    // Location tracking table
    await db.execute('''
      CREATE TABLE location_tracking (
        id TEXT PRIMARY KEY,
        user_id TEXT NOT NULL,
        job_id TEXT NOT NULL,
        location_data TEXT NOT NULL,
        type TEXT NOT NULL,
        notes TEXT,
        metadata TEXT,
        created_at TEXT,
        is_synced INTEGER DEFAULT 0
      )
    ''');

    // Routes table
    await db.execute('''
      CREATE TABLE routes (
        id TEXT PRIMARY KEY,
        start_address TEXT NOT NULL,
        end_address TEXT NOT NULL,
        start_location TEXT NOT NULL,
        end_location TEXT NOT NULL,
        distance REAL NOT NULL,
        estimated_duration REAL NOT NULL,
        waypoints TEXT,
        encoded_polyline TEXT,
        steps TEXT,
        traffic_condition TEXT,
        created_at TEXT
      )
    ''');

    // Geofence areas table
    await db.execute('''
      CREATE TABLE geofence_areas (
        id TEXT PRIMARY KEY,
        name TEXT NOT NULL,
        center_location TEXT NOT NULL,
        radius REAL NOT NULL,
        description TEXT,
        is_active INTEGER DEFAULT 1,
        trigger_events TEXT,
        metadata TEXT,
        company_id TEXT,
        created_at TEXT,
        updated_at TEXT,
        is_synced INTEGER DEFAULT 0
      )
    ''');

    // Geofence events table
    await db.execute('''
      CREATE TABLE geofence_events (
        id TEXT PRIMARY KEY,
        geofence_id TEXT NOT NULL,
        user_id TEXT NOT NULL,
        event_type TEXT NOT NULL,
        location_data TEXT NOT NULL,
        timestamp TEXT,
        job_id TEXT,
        metadata TEXT,
        created_at TEXT,
        is_synced INTEGER DEFAULT 0,
        FOREIGN KEY (geofence_id) REFERENCES geofence_areas (id) ON DELETE CASCADE
      )
    ''');

    // Sync queue table for offline operations
    await db.execute('''
      CREATE TABLE sync_queue (
        id INTEGER PRIMARY KEY AUTOINCREMENT,
        table_name TEXT NOT NULL,
        record_id TEXT NOT NULL,
        operation TEXT NOT NULL,
        data TEXT NOT NULL,
        created_at TEXT NOT NULL,
        attempts INTEGER DEFAULT 0,
        last_attempt_at TEXT,
        error_message TEXT,
        priority INTEGER DEFAULT 1
      )
    ''');

    // App settings table
    await db.execute('''
      CREATE TABLE app_settings (
        key TEXT PRIMARY KEY,
        value TEXT NOT NULL,
        updated_at TEXT NOT NULL
      )
    ''');

    // File upload queue table
    await db.execute('''
      CREATE TABLE file_upload_queue (
        id INTEGER PRIMARY KEY AUTOINCREMENT,
        file_path TEXT NOT NULL,
        upload_url TEXT,
        file_type TEXT NOT NULL,
        related_table TEXT,
        related_id TEXT,
        metadata TEXT,
        created_at TEXT NOT NULL,
        attempts INTEGER DEFAULT 0,
        last_attempt_at TEXT,
        error_message TEXT,
        is_uploaded INTEGER DEFAULT 0,
        uploaded_at TEXT
      )
    ''');
  }

  static Future<void> _createIndexes(Database db) async {
    // Job indexes
    await db.execute('CREATE INDEX idx_jobs_customer_id ON jobs (customer_id)');
    await db.execute('CREATE INDEX idx_jobs_status ON jobs (status)');
    await db.execute('CREATE INDEX idx_jobs_scheduled_date ON jobs (scheduled_date)');
    await db.execute('CREATE INDEX idx_jobs_company_id ON jobs (company_id)');
    await db.execute('CREATE INDEX idx_jobs_is_synced ON jobs (is_synced)');
    await db.execute('CREATE INDEX idx_jobs_needs_upload ON jobs (needs_upload)');

    // Job photos indexes
    await db.execute('CREATE INDEX idx_job_photos_job_id ON job_photos (job_id)');
    await db.execute('CREATE INDEX idx_job_photos_type ON job_photos (type)');
    await db.execute('CREATE INDEX idx_job_photos_needs_upload ON job_photos (needs_upload)');

    // Time entries indexes
    await db.execute('CREATE INDEX idx_time_entries_job_id ON job_time_entries (job_id)');
    await db.execute('CREATE INDEX idx_time_entries_crew_id ON job_time_entries (crew_member_id)');
    await db.execute('CREATE INDEX idx_time_entries_start_time ON job_time_entries (start_time)');

    // Equipment indexes
    await db.execute('CREATE INDEX idx_equipment_type ON equipment (type)');
    await db.execute('CREATE INDEX idx_equipment_status ON equipment (status)');
    await db.execute('CREATE INDEX idx_equipment_assigned_to ON equipment (assigned_to)');
    await db.execute('CREATE INDEX idx_equipment_company_id ON equipment (company_id)');

    // Notifications indexes
    await db.execute('CREATE INDEX idx_notifications_user_id ON notifications (user_id)');
    await db.execute('CREATE INDEX idx_notifications_type ON notifications (type)');
    await db.execute('CREATE INDEX idx_notifications_is_read ON notifications (is_read)');
    await db.execute('CREATE INDEX idx_notifications_created_at ON notifications (created_at)');

    // Location tracking indexes
    await db.execute('CREATE INDEX idx_location_user_id ON location_tracking (user_id)');
    await db.execute('CREATE INDEX idx_location_job_id ON location_tracking (job_id)');
    await db.execute('CREATE INDEX idx_location_created_at ON location_tracking (created_at)');

    // Sync queue indexes
    await db.execute('CREATE INDEX idx_sync_queue_table_record ON sync_queue (table_name, record_id)');
    await db.execute('CREATE INDEX idx_sync_queue_priority ON sync_queue (priority DESC)');
    await db.execute('CREATE INDEX idx_sync_queue_created_at ON sync_queue (created_at)');

    // File upload queue indexes
    await db.execute('CREATE INDEX idx_upload_queue_is_uploaded ON file_upload_queue (is_uploaded)');
    await db.execute('CREATE INDEX idx_upload_queue_related ON file_upload_queue (related_table, related_id)');
    await db.execute('CREATE INDEX idx_upload_queue_created_at ON file_upload_queue (created_at)');
  }

  static Future<void> clearAllTables(Database db) async {
    await db.execute('DELETE FROM file_upload_queue');
    await db.execute('DELETE FROM sync_queue');
    await db.execute('DELETE FROM geofence_events');
    await db.execute('DELETE FROM geofence_areas');
    await db.execute('DELETE FROM routes');
    await db.execute('DELETE FROM location_tracking');
    await db.execute('DELETE FROM notifications');
    await db.execute('DELETE FROM equipment_usage');
    await db.execute('DELETE FROM maintenance_records');
    await db.execute('DELETE FROM equipment');
    await db.execute('DELETE FROM services');
    await db.execute('DELETE FROM job_time_entries');
    await db.execute('DELETE FROM job_photos');
    await db.execute('DELETE FROM jobs');
  }

  static Future<void> resetDatabase(Database db) async {
    await clearAllTables(db);
    // Reset any auto-increment counters
    await db.execute('DELETE FROM sqlite_sequence');
  }
}