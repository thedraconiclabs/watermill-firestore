package firestore

import (
	"context"

	"cloud.google.com/go/firestore"
	"github.com/pkg/errors"
	"google.golang.org/api/option"
)

// client defines all the methods original client implements.
// It's meant to be used everywhere client is needed in order to enable decorating the original implementation.
type client interface {
	RunTransaction(
		ctx context.Context,
		f func(context.Context, *firestore.Transaction) error,
		opts ...firestore.TransactionOption,
	) (err error)
	Collection(path string) *firestore.CollectionRef
	Doc(path string) *firestore.DocumentRef
	CollectionGroup(collectionID string) *firestore.CollectionGroupRef
	GetAll(ctx context.Context, docRefs []*firestore.DocumentRef) ([]*firestore.DocumentSnapshot, error)
	Collections(ctx context.Context) *firestore.CollectionIterator
	//nolint we need to cover entire interface, even deprecated methods
	Batch() *firestore.WriteBatch
	Close() error
}

// resolveClient builds (or reuses) the Firestore client used by a publisher or
// subscriber, applying the fork's client-injection and named-database rules:
//
//   - If customClient is non-nil (an injected implementation of the internal
//     client interface), it is returned as-is.
//   - If firestoreClient is non-nil (an injected *firestore.Client), it is
//     returned as-is.
//   - Otherwise a new *firestore.Client is created. When database is non-empty
//     it selects a named database via firestore.NewClientWithDatabase;
//     otherwise it uses the default database via firestore.NewClient.
func resolveClient(
	customClient client,
	firestoreClient *firestore.Client,
	projectID string,
	database string,
	opts []option.ClientOption,
) (client, error) {
	if customClient != nil {
		return customClient, nil
	}
	if firestoreClient != nil {
		return firestoreClient, nil
	}

	if database != "" {
		c, err := firestore.NewClientWithDatabase(context.Background(), projectID, database, opts...)
		if err != nil {
			return nil, errors.Wrap(err, "cannot create firestore client for named database")
		}
		return c, nil
	}

	c, err := firestore.NewClient(context.Background(), projectID, opts...)
	if err != nil {
		return nil, errors.Wrap(err, "cannot create default firestore client")
	}
	return c, nil
}
