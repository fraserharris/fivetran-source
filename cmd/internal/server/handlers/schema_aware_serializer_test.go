package handlers

import (
	"math"
	"testing"
	"time"

	"github.com/planetscale/fivetran-source/lib"

	"google.golang.org/protobuf/types/known/timestamppb"

	fivetransdk "github.com/planetscale/fivetran-sdk-grpc/go"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"vitess.io/vitess/go/sqltypes"
	querypb "vitess.io/vitess/go/vt/proto/query"
)

func TestCanSerializeInsert(t *testing.T) {
	timestamp := "2006-01-02 15:04:05"
	row, s, err := generateTestRecord("PhaniRaj")
	require.NoError(t, err)
	tl := &testLogSender{}
	l := NewSchemaAwareSerializer(tl, "", true, &fivetransdk.SchemaList{Schemas: []*fivetransdk.Schema{s}})

	schema := &fivetransdk.SchemaSelection{
		Included:   true,
		SchemaName: s.Name,
	}
	table := &fivetransdk.TableSelection{
		TableName: "Customers",
		Included:  true,
		Columns:   map[string]bool{},
	}

	for _, f := range row.Fields {
		table.Columns[f.Name] = true
	}

	for i := 0; i < 3; i++ {
		err = l.Record(row, schema, table, lib.OpType_Insert)
		assert.NoError(t, err)
		assert.NotNil(t, tl.lastResponse)
	}

	operation, ok := tl.lastResponse.Response.(*fivetransdk.UpdateResponse_Operation)
	require.Truef(t, ok, "recordResponse Operation is not of type %s", "UpdateResponse_Operation")

	operationRecord, ok := operation.Operation.Op.(*fivetransdk.Operation_Record)
	assert.Truef(t, ok, "recordResponse Operation.Op is not of type %s", "Operation_Record")

	assert.Equal(t, fivetransdk.OpType_UPSERT, operationRecord.Record.Type)
	data := operationRecord.Record.Data
	assert.NotNil(t, data)

	assert.Equal(t, int32(123), data["customer_id"].GetInt())
	assert.Equal(t, "string:\"PhaniRaj\"", data["name"].String())
	assert.Equal(t, "string:\"Something great comes this way\"", data["notes"].String())
	assert.False(t, data["is_deleted"].GetBool())
	assert.Equal(t, "156.123", data["decimal"].GetDecimal())
	assert.Equal(t, []byte("profiles/phanatic.jpg"), data["profile_pic"].GetBinary())
	assert.Equal(t, "{'home': 'phanatic.dev'}", data["sitemap"].GetJson())
	assert.Equal(t, int64(math.MaxInt64), data["long_value"].GetLong())
	assert.Equal(t, math.MaxFloat64, data["double_value"].GetDouble())
	assert.Equal(t, float32(123.456), data["float_value"].GetFloat())
	dateValue, err := time.Parse("2006-01-02", "2004-12-12")
	require.NoError(t, err)
	assert.Equal(t, timestamppb.New(dateValue).Nanos, data["date_value"].GetNaiveDate().Nanos)
	dts, err := time.Parse(timestamp, "2006-01-02 15:04:05")
	require.NoError(t, err)
	assert.Equal(t, timestamppb.New(dts).Nanos, data["timestamp_value"].GetUtcDatetime().Nanos)
	dt, err := time.Parse("2006-01-02 15:04:05", "2021-01-19 03:14:07.999999")
	require.NoError(t, err)
	assert.Equal(t, timestamppb.New(dt).Nanos, data["datetime_value"].GetNaiveDatetime().Nanos)
	assert.True(t, data["tiny_int_as_bool_value"].GetBool())
}

func TestCanSerializeNulLValues(t *testing.T) {
	row, s, err := generateTestRecord("PhaniRaj")
	require.NoError(t, err)
	row.Rows = [][]sqltypes.Value{
		{
			sqltypes.NULL,
			sqltypes.NULL,
			sqltypes.NULL,
			sqltypes.NULL,
			sqltypes.NULL,
			sqltypes.NULL,
			sqltypes.NULL,
			sqltypes.NULL,
			sqltypes.NULL,
			sqltypes.NULL,
			sqltypes.NULL,
			sqltypes.NULL,
			sqltypes.NULL,
			sqltypes.NULL,
			sqltypes.NULL,
			sqltypes.NULL,
			sqltypes.NULL,
			sqltypes.NULL,
		},
	}

	tl := &testLogSender{}
	l := NewSchemaAwareSerializer(tl, "", false, &fivetransdk.SchemaList{Schemas: []*fivetransdk.Schema{s}})
	schema := &fivetransdk.SchemaSelection{
		Included:   true,
		SchemaName: s.Name,
	}
	table := &fivetransdk.TableSelection{
		TableName: "Customers",
		Included:  true,
		Columns:   map[string]bool{},
	}

	for _, f := range row.Fields {
		table.Columns[f.Name] = true
	}

	for i := 0; i < 3; i++ {
		err = l.Record(row, schema, table, lib.OpType_Delete)
		assert.NoError(t, err)
		assert.NotNil(t, tl.lastResponse)
	}

	operation, ok := tl.lastResponse.Response.(*fivetransdk.UpdateResponse_Operation)
	require.Truef(t, ok, "recordResponse Operation is not of type %s", "UpdateResponse_Operation")

	operationRecord, ok := operation.Operation.Op.(*fivetransdk.Operation_Record)
	assert.Truef(t, ok, "recordResponse Operation.Op is not of type %s", "Operation_Record")

	assert.Equal(t, fivetransdk.OpType_DELETE, operationRecord.Record.Type)
	data := operationRecord.Record.Data
	assert.NotNil(t, data)
	assert.Equal(t, 2, len(data), "should serialize only primary keys for deleted rows")
	for _, value := range data {
		inner := value.Inner
		assert.IsType(t, &fivetransdk.ValueType_Null{}, inner)
		assert.True(t, inner.(*fivetransdk.ValueType_Null).Null)
	}
}

func TestCanSerializeDelete(t *testing.T) {
	row, s, err := generateTestRecord("PhaniRaj")
	require.NoError(t, err)
	tl := &testLogSender{}
	l := NewSchemaAwareSerializer(tl, "", false, &fivetransdk.SchemaList{Schemas: []*fivetransdk.Schema{s}})

	schema := &fivetransdk.SchemaSelection{
		Included:   true,
		SchemaName: s.Name,
	}
	table := &fivetransdk.TableSelection{
		TableName: "Customers",
		Included:  true,
		Columns:   map[string]bool{},
	}

	for _, f := range row.Fields {
		table.Columns[f.Name] = true
	}

	for i := 0; i < 3; i++ {
		err = l.Record(row, schema, table, lib.OpType_Delete)
		assert.NoError(t, err)
		assert.NotNil(t, tl.lastResponse)
	}

	operation, ok := tl.lastResponse.Response.(*fivetransdk.UpdateResponse_Operation)
	require.Truef(t, ok, "recordResponse Operation is not of type %s", "UpdateResponse_Operation")

	operationRecord, ok := operation.Operation.Op.(*fivetransdk.Operation_Record)
	assert.Truef(t, ok, "recordResponse Operation.Op is not of type %s", "Operation_Record")

	assert.Equal(t, fivetransdk.OpType_DELETE, operationRecord.Record.Type)
	data := operationRecord.Record.Data
	assert.NotNil(t, data)
	assert.Equal(t, 2, len(data), "should serialize only primary keys for deleted rows")
	assert.Equal(t, int32(123), data["customer_id"].GetInt())
	assert.Equal(t, "string:\"PhaniRaj\"", data["name"].String())
}

func TestCanSerializeUpdate(t *testing.T) {
	before, s, err := generateTestRecord("PhaniRaj")

	require.NoError(t, err)
	tl := &testLogSender{}
	l := NewSchemaAwareSerializer(tl, "", false, &fivetransdk.SchemaList{Schemas: []*fivetransdk.Schema{s}})

	schema := &fivetransdk.SchemaSelection{
		Included:   true,
		SchemaName: s.Name,
	}
	table := &fivetransdk.TableSelection{
		TableName: "Customers",
		Included:  true,
		Columns:   map[string]bool{},
	}

	for _, f := range before.Fields {
		table.Columns[f.Name] = true
	}

	after, _, err := generateTestRecord("YayavaramNarasimha")
	assert.NoError(t, err)

	for i := 0; i < 3; i++ {
		err = l.Update(&lib.UpdatedRow{
			Before: before,
			After:  after,
		}, schema, table)

		assert.NoError(t, err)
		assert.NotNil(t, tl.lastResponse)
	}

	operation, ok := tl.lastResponse.Response.(*fivetransdk.UpdateResponse_Operation)
	require.Truef(t, ok, "recordResponse Operation is not of type %s", "UpdateResponse_Operation")

	operationRecord, ok := operation.Operation.Op.(*fivetransdk.Operation_Record)
	assert.Truef(t, ok, "recordResponse Operation.Op is not of type %s", "Operation_Record")

	assert.Equal(t, fivetransdk.OpType_UPDATE, operationRecord.Record.Type)
	data := operationRecord.Record.Data
	assert.NotNil(t, data)
	assert.Equal(t, 2, len(data))
	assert.Equal(t, int32(123), data["customer_id"].GetInt())
	assert.Equal(t, "string:\"YayavaramNarasimha\"", data["name"].String())
}

func TestCanSerializeTruncate(t *testing.T) {
	_, s, err := generateTestRecord("PhaniRaj")
	assert.NoError(t, err)
	tl := &testLogSender{}
	l := NewSchemaAwareSerializer(tl, "", false, &fivetransdk.SchemaList{Schemas: []*fivetransdk.Schema{s}})

	schema := &fivetransdk.SchemaSelection{
		Included:   true,
		SchemaName: s.Name,
	}
	table := &fivetransdk.TableSelection{
		TableName: "Customers",
		Included:  true,
		Columns:   map[string]bool{},
	}

	for i := 0; i < 3; i++ {
		err = l.Truncate(schema, table)

		assert.NoError(t, err)
		assert.NotNil(t, tl.lastResponse)
	}

	operation, ok := tl.lastResponse.Response.(*fivetransdk.UpdateResponse_Operation)
	require.Truef(t, ok, "recordResponse Operation is not of type %s", "UpdateResponse_Operation")

	operationRecord, ok := operation.Operation.Op.(*fivetransdk.Operation_Record)
	assert.Truef(t, ok, "recordResponse Operation.Op is not of type %s", "Operation_Record")

	assert.Equal(t, fivetransdk.OpType_TRUNCATE, operationRecord.Record.Type)
	assert.Nil(t, operationRecord.Record.Data)
}

func generateTestRecord(name string) (*sqltypes.Result, *fivetransdk.Schema, error) {
	notes, err := sqltypes.NewValue(querypb.Type_TEXT, []byte("Something great comes this way"))
	if err != nil {
		return nil, nil, err
	}
	decimal, err := sqltypes.NewValue(querypb.Type_DECIMAL, []byte("156.123"))
	if err != nil {
		return nil, nil, err
	}
	profilePic, err := sqltypes.NewValue(querypb.Type_BINARY, []byte("profiles/phanatic.jpg"))
	if err != nil {
		return nil, nil, err
	}
	siteMap, err := sqltypes.NewValue(querypb.Type_JSON, []byte("{'home': 'phanatic.dev'}"))
	if err != nil {
		return nil, nil, err
	}
	floatValue, err := sqltypes.NewValue(querypb.Type_FLOAT32, []byte("123.456"))
	if err != nil {
		return nil, nil, err
	}
	timestamp := "2006-01-02 15:04:05"
	row := &sqltypes.Result{
		Fields: []*querypb.Field{
			{
				Name:         "customer_id",
				Type:         querypb.Type_INT32,
				ColumnLength: 21,
				Charset:      63,
				Flags:        32928,
			},
			{
				Name:         "name",
				Type:         querypb.Type_VARCHAR,
				ColumnLength: 21,
				Charset:      63,
				Flags:        32928,
			},
			{
				Name:         "first_name",
				Type:         querypb.Type_VARCHAR,
				ColumnLength: 21,
				Charset:      63,
				Flags:        32928,
			},
			{
				Name:         "middle_name",
				Type:         querypb.Type_VARCHAR,
				ColumnLength: 21,
				Charset:      63,
				Flags:        32928,
			},
			{
				Name:         "last_name",
				Type:         querypb.Type_VARCHAR,
				ColumnLength: 21,
				Charset:      63,
				Flags:        32928,
			},
			{
				Name:  "is_deleted",
				Type:  querypb.Type_INT8,
				Flags: uint32(querypb.Flag_ISINTEGRAL),
			},
			{
				Name:    "notes",
				Type:    querypb.Type_TEXT,
				Charset: 63,
				Flags:   32928,
			},
			{
				Name:    "decimal",
				Type:    querypb.Type_DECIMAL,
				Charset: 63,
				Flags:   32928,
			},
			{
				Name:  "profile_pic",
				Type:  querypb.Type_BINARY,
				Flags: 32928,
			},
			{
				Name:  "header_pic",
				Type:  querypb.Type_BINARY,
				Flags: 32928,
			},
			{
				Name:  "footer_pic",
				Type:  querypb.Type_BINARY,
				Flags: 32928,
			},
			{
				Name:  "sitemap",
				Type:  querypb.Type_JSON,
				Flags: 32928,
			},
			{
				Name:  "long_value",
				Type:  querypb.Type_INT64,
				Flags: 32928,
			},
			{
				Name:  "double_value",
				Type:  querypb.Type_FLOAT64,
				Flags: 32928,
			},
			{
				Name:  "float_value",
				Type:  querypb.Type_FLOAT32,
				Flags: 32928,
			},
			{
				Name:  "date_value",
				Type:  querypb.Type_DATE,
				Flags: 32928,
			},
			{
				Name:  "timestamp_value",
				Type:  querypb.Type_TIMESTAMP,
				Flags: 32928,
			},
			{
				Name:  "datetime_value",
				Type:  querypb.Type_DATETIME,
				Flags: 32928,
			},
			{
				Name:  "tiny_int_as_bool_value",
				Type:  querypb.Type_INT8,
				Flags: 32928,
			},
		},
		Rows: [][]sqltypes.Value{
			{
				sqltypes.NewInt32(123),
				sqltypes.NewVarChar(name),
				sqltypes.NewVarChar("PhaniRaj"),
				sqltypes.NewVarChar("PhaniRaj"),
				sqltypes.NewVarChar("PhaniRaj"),
				sqltypes.NewInt8(0),
				notes,
				decimal,
				profilePic,
				profilePic,
				profilePic,
				siteMap,
				sqltypes.NewInt64(math.MaxInt64),
				sqltypes.NewFloat64(math.MaxFloat64),
				floatValue,
				sqltypes.NewDate("2004-12-12"),
				sqltypes.NewTimestamp(timestamp),
				sqltypes.NewDatetime("2021-01-19 03:14:07.999999"),
				sqltypes.NewInt32(1),
			},
		},
	}
	return row, &fivetransdk.Schema{
		Name: "sample",
		Tables: []*fivetransdk.Table{
			{
				Name: "Customers",
				Columns: []*fivetransdk.Column{
					{
						Name:       "customer_id",
						Type:       fivetransdk.DataType_INT,
						PrimaryKey: true,
					},
					{
						Name:       "name",
						Type:       fivetransdk.DataType_STRING,
						PrimaryKey: true,
					},
					{
						Name: "first_name",
						Type: fivetransdk.DataType_STRING,
					},
					{
						Name: "last_name",
						Type: fivetransdk.DataType_STRING,
					},
					{
						Name: "middle_name",
						Type: fivetransdk.DataType_STRING,
					},
					{
						Name: "is_deleted",
						Type: fivetransdk.DataType_BOOLEAN,
					},
					{
						Name: "notes",
						Type: fivetransdk.DataType_STRING,
					},
					{
						Name: "decimal",
						Type: fivetransdk.DataType_DECIMAL,
					},
					{
						Name: "profile_pic",
						Type: fivetransdk.DataType_BINARY,
					},
					{
						Name: "header_pic",
						Type: fivetransdk.DataType_BINARY,
					},
					{
						Name: "footer_pic",
						Type: fivetransdk.DataType_BINARY,
					},
					{
						Name: "sitemap",
						Type: fivetransdk.DataType_JSON,
					},
					{
						Name: "long_value",
						Type: fivetransdk.DataType_LONG,
					},
					{
						Name: "double_value",
						Type: fivetransdk.DataType_DOUBLE,
					},
					{
						Name: "float_value",
						Type: fivetransdk.DataType_FLOAT,
					},
					{
						Name: "date_value",
						Type: fivetransdk.DataType_NAIVE_DATE,
					},
					{
						Name: "timestamp_value",
						Type: fivetransdk.DataType_UTC_DATETIME,
					},
					{
						Name: "datetime_value",
						Type: fivetransdk.DataType_NAIVE_DATETIME,
					},
					{
						Name: "tiny_int_as_bool_value",
						Type: fivetransdk.DataType_BOOLEAN,
					},
				},
			},
		},
	}, err
}

func TestCanSkipColumns(t *testing.T) {
	row := &sqltypes.Result{
		Fields: []*querypb.Field{
			{
				Name:         "customer_id",
				Type:         querypb.Type_INT32,
				ColumnLength: 21,
				Charset:      63,
				Flags:        32928,
			},
			{
				Name:         "name",
				Type:         querypb.Type_VARCHAR,
				ColumnLength: 21,
				Charset:      63,
				Flags:        32928,
			},
		},
		Rows: [][]sqltypes.Value{
			{
				sqltypes.NewInt32(123),
				sqltypes.NewVarChar("PhaniRaj"),
			},
		},
	}

	tl := &testLogSender{}
	l := NewSchemaAwareSerializer(tl, "", false, &fivetransdk.SchemaList{Schemas: []*fivetransdk.Schema{
		{
			Name: "SalesDB",
			Tables: []*fivetransdk.Table{
				{
					Name: "Customers",
					Columns: []*fivetransdk.Column{
						{
							Name: "customer_id",
							Type: fivetransdk.DataType_INT,
						},
						{
							Name: "name",
							Type: fivetransdk.DataType_STRING,
						},
					},
				},
			},
		},
	}})

	schema := &fivetransdk.SchemaSelection{
		Included:   true,
		SchemaName: "SalesDB",
	}
	table := &fivetransdk.TableSelection{
		TableName: "Customers",
		Included:  true,
		Columns: map[string]bool{
			"customer_id": true,
			"name":        false,
		},
	}

	err := l.Record(row, schema, table, 0)
	assert.NoError(t, err)
	assert.NotNil(t, tl.lastResponse)

	operation, ok := tl.lastResponse.Response.(*fivetransdk.UpdateResponse_Operation)
	require.Truef(t, ok, "recordResponse Operation is not of type %s", "UpdateResponse_Operation")

	operationRecord, ok := operation.Operation.Op.(*fivetransdk.Operation_Record)
	assert.Truef(t, ok, "recordResponse Operation.Op is not of type %s", "Operation_Record")

	data := operationRecord.Record.Data
	assert.NotNil(t, data)

	assert.Equal(t, int32(123), data["customer_id"].GetInt())
	_, found := data["name"]
	assert.False(t, found, "should not include unselected column in output")
}

func BenchmarkRecordSerialization_Serializer(b *testing.B) {
	row, s, err := generateTestRecord("PhaniRaj")
	if err != nil {
		panic(err.Error())
	}

	tl := &testLogSender{}
	l := NewSchemaAwareSerializer(tl, "", false, &fivetransdk.SchemaList{
		Schemas: []*fivetransdk.Schema{
			s,
		},
	})

	schema := &fivetransdk.SchemaSelection{
		SchemaName: "SalesDB",
	}
	table := &fivetransdk.TableSelection{
		TableName: "Customers",
	}

	for n := 0; n < b.N; n++ {
		err := l.Record(row, schema, table, 0)
		if err != nil {
			b.Fatalf("failed with %v", err.Error())
		}
	}
}
