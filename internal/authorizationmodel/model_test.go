package authorizationmodel_test

import (
	"math"
	"strings"
	"testing"

	openfga "github.com/openfga/go-sdk"
	"github.com/openfga/openfga/pkg/typesystem"
	"google.golang.org/protobuf/proto"

	"github.com/openfga/cli/internal/authorizationmodel"
)

const (
	modelID        = "01GVKXGDCV2SMG6TRE9NMBQ2VG"
	typeName       = "user"
	modelCreatedAt = "2023-03-16 00:35:51.835 +0000 UTC"
)

func TestReadingInvalidModelFromInvalidJSON(t *testing.T) {
	t.Parallel()

	modelString := "{bad_json"

	model := authorizationmodel.AuthzModel{}

	err := model.ReadFromJSONString(modelString)
	if err == nil {
		t.Errorf("Expected error, got none")
	}
}

func TestReadingValidModelFromJSON(t *testing.T) {
	t.Parallel()

	modelString := `{"id":"01GVKXGDCV2SMG6TRE9NMBQ2VG","schema_version":"1.1","type_definitions":[{"type":"user"}]}`

	model := authorizationmodel.AuthzModel{}

	err := model.ReadFromJSONString(modelString)
	if err != nil {
		t.Errorf("Got error when reading a valid model %v", err)
	}

	if model.GetSchemaVersion() != typesystem.SchemaVersion1_1 {
		t.Errorf("Expected %v to equal %v", model.GetSchemaVersion(), typesystem.SchemaVersion1_1)
	}

	if model.GetID() != modelID {
		t.Errorf("Expected %v to equal %v", model.GetID(), modelID)
	}

	if model.CreatedAt.String() != modelCreatedAt {
		t.Errorf("Expected %v to equal %v", model.GetCreatedAt().String(), modelCreatedAt)
	}

	if model.GetTypeDefinitions()[0].GetType() != typeName {
		t.Errorf("Expected %v to equal %v", model.GetTypeDefinitions()[0].GetType(), typeName)
	}
}

func TestReadingValidModelFromDSL(t *testing.T) {
	t.Parallel()

	model := authorizationmodel.AuthzModel{}
	if err := model.ReadFromDSLString(`model
  schema 1.1

type user
`); err != nil {
		t.Errorf("Got error when parsing a valid model %v", err)
	}

	if model.GetSchemaVersion() != typesystem.SchemaVersion1_1 {
		t.Errorf("Expected %v to equal %v", model.GetSchemaVersion(), typesystem.SchemaVersion1_1)
	}

	if model.GetTypeDefinitions()[0].GetType() != typeName {
		t.Errorf("Expected %v to equal %v", model.GetTypeDefinitions()[0].GetType(), typeName)
	}
}

func TestDisplayAsJsonWithFields(t *testing.T) {
	t.Parallel()

	typeDefs := []openfga.TypeDefinition{{
		Type: typeName,
	}}
	model := authorizationmodel.AuthzModel{
		SchemaVersion:   openfga.PtrString(typesystem.SchemaVersion1_1),
		ID:              openfga.PtrString(modelID),
		TypeDefinitions: &typeDefs,
	}

	jsonModel1 := model.DisplayAsJSON([]string{"model", "id", "created_at"})
	if jsonModel1.GetSchemaVersion() != typesystem.SchemaVersion1_1 {
		t.Errorf("Expected %v to equal %v", jsonModel1.GetSchemaVersion(), typesystem.SchemaVersion1_1)
	}

	if jsonModel1.GetID() != modelID {
		t.Errorf("Expected %v to equal %v", jsonModel1.GetID(), modelID)
	}

	if jsonModel1.GetCreatedAt().String() != modelCreatedAt {
		t.Errorf("Expected %v to equal %v", jsonModel1.GetCreatedAt().String(), modelCreatedAt)
	}

	if jsonModel1.GetTypeDefinitions()[0].GetType() != typeName {
		t.Errorf("Expected %v to equal %v", jsonModel1.GetTypeDefinitions()[0].GetType(), typeName)
	}

	jsonModel2 := model.DisplayAsJSON([]string{"id", "created_at"})
	if jsonModel2.GetSchemaVersion() != "" {
		t.Errorf("Expected %v to be empty", jsonModel2.GetSchemaVersion())
	}

	if jsonModel2.GetID() != modelID {
		t.Errorf("Expected %v to equal %v", jsonModel2.GetID(), modelID)
	}

	if jsonModel1.GetCreatedAt().String() != modelCreatedAt {
		t.Errorf("Expected %v to equal %v", jsonModel2.GetCreatedAt().String(), modelCreatedAt)
	}

	if jsonModel2.GetTypeDefinitions() != nil {
		t.Errorf("Expected %v to equal nil", jsonModel2.GetTypeDefinitions())
	}
}

func TestGetSizeInKB(t *testing.T) {
	t.Parallel()

	typeDefs := []openfga.TypeDefinition{{Type: typeName}}
	model := authorizationmodel.AuthzModel{
		SchemaVersion:   openfga.PtrString(typesystem.SchemaVersion1_1),
		ID:              openfga.PtrString(modelID),
		TypeDefinitions: &typeDefs,
	}

	size := model.GetSizeInKB()
	if size <= 0 {
		t.Errorf("Expected positive size, got %v", size)
	}

	pbModel := model.GetProtoModel()

	bytes, err := proto.Marshal(pbModel)
	if err != nil {
		t.Fatalf("unexpected marshal error: %v", err)
	}

	expected := math.Round(float64(len(bytes))/1024.0*100) / 100
	if size != expected {
		t.Errorf("Expected %v to equal %v", size, expected)
	}

	if size != math.Round(size*100)/100 {
		t.Errorf("Expected %v to be rounded to 2 decimal places", size)
	}
}

func TestGetSizeInKBWithCreatedAt(t *testing.T) {
	t.Parallel()

	model := authorizationmodel.AuthzModel{}
	model.Set(openfga.AuthorizationModel{
		Id:              modelID,
		SchemaVersion:   typesystem.SchemaVersion1_1,
		TypeDefinitions: []openfga.TypeDefinition{{Type: typeName}},
	})

	if model.GetCreatedAt() == nil {
		t.Fatalf("expected CreatedAt to be populated by Set")
	}

	size := model.GetSizeInKB()
	if size <= 0 {
		t.Errorf("Expected positive size for store-fetched model, got %v", size)
	}
}

func TestDisplayAsJSONWithSize(t *testing.T) {
	t.Parallel()

	typeDefs := []openfga.TypeDefinition{{Type: typeName}}
	model := authorizationmodel.AuthzModel{
		SchemaVersion:   openfga.PtrString(typesystem.SchemaVersion1_1),
		ID:              openfga.PtrString(modelID),
		TypeDefinitions: &typeDefs,
	}

	withSize := model.DisplayAsJSON([]string{"model", "size"})
	if withSize.SizeKB == nil {
		t.Fatalf("Expected SizeKB to be set")
	}

	if *withSize.SizeKB != model.GetSizeInKB() {
		t.Errorf("Expected %v to equal %v", *withSize.SizeKB, model.GetSizeInKB())
	}

	withoutSize := model.DisplayAsJSON([]string{"model"})
	if withoutSize.SizeKB != nil {
		t.Errorf("Expected SizeKB to be nil, got %v", *withoutSize.SizeKB)
	}
}

func TestDisplayAsDSLWithSize(t *testing.T) {
	t.Parallel()

	model := authorizationmodel.AuthzModel{}
	if err := model.ReadFromDSLString("model\n  schema 1.1\n\ntype user\n"); err != nil {
		t.Fatalf("unexpected parse error: %v", err)
	}

	withSize, err := model.DisplayAsDSL([]string{"model", "size"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(*withSize, "# Size: ") {
		t.Errorf("Expected DSL output to contain a size comment, got %v", *withSize)
	}

	withoutSize, err := model.DisplayAsDSL([]string{"model"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if strings.Contains(*withoutSize, "# Size: ") {
		t.Errorf("Expected DSL output to omit size comment, got %v", *withoutSize)
	}
}
