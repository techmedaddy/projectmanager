// Package projects contains project-related storage and business logic.
//
// Project visibility rules:
//   - owners can list, read, update, and delete their own projects
//   - non-owners can list and read a project only when they currently have at
//     least one task assigned in that project
//   - non-owners cannot update or delete projects
//
// These rules intentionally make "project visibility" stricter than "task
// creation history": creating a task in the past does not keep a project
// visible forever unless the user still owns the project or has an assigned
// task in it.
package projects
