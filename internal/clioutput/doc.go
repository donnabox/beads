// Package clioutput projects and renders already-selected Resources,
// collections, and complete mutation results.
// It has no storage, transport, command registration, or process-output side
// effects. Adapters supply the exact BDP contract and protection descriptor;
// this package never infers Resource kinds from user properties or fetches
// missing fields. Collection/streaming and command compatibility are separate
// adapter work.
package clioutput
