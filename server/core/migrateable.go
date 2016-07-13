package core

import (
	m "github.com/blendlabs/spiffy/migration"
)

// Migrateable is a type that provides a migration.
type Migrateable interface {
	Migration() m.Migration
}
