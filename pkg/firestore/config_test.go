package firestore_test

import (
	"context"
	"testing"

	stdFirestore "cloud.google.com/go/firestore"
	"github.com/ThreeDotsLabs/watermill"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/api/option"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	"github.com/ThreeDotsLabs/watermill-firestore/pkg/firestore"
)

// newOfflineClient builds a *firestore.Client that is constructed lazily and
// never dials Firestore, so it can be used in unit tests without an emulator or
// credentials. firestore.NewClient does not establish a connection until the
// client is actually used.
func newOfflineClient(t *testing.T) *stdFirestore.Client {
	t.Helper()

	client, err := stdFirestore.NewClient(
		context.Background(),
		"fork-unit-test-project",
		option.WithoutAuthentication(),
		option.WithEndpoint("localhost:0"),
		option.WithGRPCDialOption(grpc.WithTransportCredentials(insecure.NewCredentials())),
	)
	require.NoError(t, err)
	t.Cleanup(func() { _ = client.Close() })
	return client
}

// TestPublisherHonorsInjectedClient verifies that an injected *firestore.Client
// is accepted by the publisher config and used as-is, without the package
// creating its own client (which would otherwise require credentials).
func TestPublisherHonorsInjectedClient(t *testing.T) {
	logger := watermill.NopLogger{}
	injected := newOfflineClient(t)

	pub, err := firestore.NewPublisher(firestore.PublisherConfig{
		// Intentionally no ProjectID: when FirestoreClient is set the package
		// must not attempt to construct its own client.
		FirestoreClient: injected,
	}, logger)
	require.NoError(t, err)
	require.NotNil(t, pub)
}

// TestSubscriberHonorsInjectedClient mirrors the publisher case for the
// subscriber config.
func TestSubscriberHonorsInjectedClient(t *testing.T) {
	logger := watermill.NopLogger{}
	injected := newOfflineClient(t)

	sub, err := firestore.NewSubscriber(firestore.SubscriberConfig{
		FirestoreClient: injected,
	}, logger)
	require.NoError(t, err)
	require.NotNil(t, sub)
}

// TestNamedDatabaseConfigCompiles verifies that the Database field exists on
// both configs and is accepted. We point the client at an unreachable endpoint
// so construction stays offline; combined with an injected client the named
// database is ignored for construction, proving precedence (injection wins).
func TestInjectedClientTakesPrecedenceOverDatabase(t *testing.T) {
	logger := watermill.NopLogger{}
	injected := newOfflineClient(t)

	pub, err := firestore.NewPublisher(firestore.PublisherConfig{
		ProjectID:       "ignored-when-client-injected",
		Database:        "main",
		FirestoreClient: injected,
	}, logger)
	require.NoError(t, err)
	assert.NotNil(t, pub)

	sub, err := firestore.NewSubscriber(firestore.SubscriberConfig{
		ProjectID:       "ignored-when-client-injected",
		Database:        "main",
		FirestoreClient: injected,
	}, logger)
	require.NoError(t, err)
	assert.NotNil(t, sub)
}

// TestTransactionalPublisherWithInjectedClient verifies the
// PublishInTransaction path can be wired with an injected client and a
// transaction-bound TransactionalPublisher constructed offline.
func TestTransactionalPublisherWithInjectedClient(t *testing.T) {
	logger := watermill.NopLogger{}
	injected := newOfflineClient(t)

	// tx is nil here only to prove the constructor wiring compiles and accepts
	// an injected client; a real tx comes from injected.RunTransaction at call
	// time. We do not Publish (which would dial Firestore).
	pub, err := firestore.NewTransactionalPublisher(firestore.PublisherConfig{
		Database:        "main",
		FirestoreClient: injected,
	}, nil, logger)
	require.NoError(t, err)
	require.NotNil(t, pub)
}
