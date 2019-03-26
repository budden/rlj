package leftjoin

import (
	"fmt"
	"time"

	"github.com/go-redis/redis"
	"github.com/pkg/errors"
)

const zeroDuration = time.Duration(0)

// ObtainAnSetNxLockForTable tries to lock table, using SETNX with key = "tablename.updating"
func ObtainAnSetNxLockForTable(rc *redis.Client,
	tableName string,
	timeout, retryPeriod time.Duration) (obtainedALock bool, err error) {

	redisKeyForLock := tableName + ".updating"

	tryOnce := func() (returnNow bool) {
		obtainedALock, err = rc.SetNX(redisKeyForLock, 1, time.Duration(0)).Result()
		if err != nil {
			err = errors.Wrapf(err, "Error locking %s", tableName)
		}
		returnNow = err != nil || obtainedALock
		return
	}

	if tryOnce() {
		return
	}

	// ok, let's wait now
	if timeout < 0 || retryPeriod < 0 {
		err = errors.New("timeout and retryPeriod must be nonnegative")
		return
	}
	if timeout != zeroDuration &&
		retryPeriod != zeroDuration &&
		retryPeriod > timeout {
		retryPeriod = timeout
	}

	expirationTime := time.Now().Local().Add(timeout)
	for {
		if err != nil || obtainedALock {
			return
		}

		if retryPeriod != zeroDuration {
			time.Sleep(retryPeriod)
		}
		if tryOnce() {
			return
		}
		if time.Now().Local().After(expirationTime) {
			err = fmt.Errorf("Locking of %s table timed out", tableName)
			return
		}
	}
}

// ReleaseAnSetNxLockForTable releases a lock. If lock was not there, error is returned
func ReleaseAnSetNxLockForTable(rc *redis.Client, tableName string) (err error) {
	redisKeyForLock := tableName + ".updating"
	_, err = rc.Get(redisKeyForLock).Result()
	if err == redis.Nil {
		err = fmt.Errorf("Lock protocol violation: while releasing lock on table %s, lock was not set", tableName)
		return
	} else if err != nil {
		return
	}
	var count int64
	count, err = rc.Del(redisKeyForLock).Result()
	if err != nil {
		return
	}
	if count != 1 {
		err = fmt.Errorf("Lock protocol violation: del returned %d when unlockin table %s, race condition suspected", count, tableName)
	}
	return
}

/* // LockError is used to report lock errors, it can encapsulate another error
type LockError struct {
	Err *error
	TableName string
	// mode is either "lock" or "unlock"
	Mode string
}

// Error implements errors.Error()
func (e* LockError) Error() string {
	return fmt.Sprint(e)
} */

// WithNxLock obtains a lock on a table, then runs body, then, if body returns no error, releases a lock
// Lock is implemented with the SETNX $tableName.updating command. While obtaining, WithNxLock waits for
// 3 seconds and then returns an error with "Failed to lock table %s" message
func WithNxLock(rc *redis.Client, tableName string, body func() error) (err error) {
	// We modify several coordinated structures, so we need a sort of lock
	var obtainedALock bool
	obtainedALock, err = ObtainAnSetNxLockForTable(rc, tableName, time.Second*3, time.Millisecond*100)

	if err != nil {
		err = errors.Wrapf(err, "Failed to lock table orders")
		return
	}

	if !obtainedALock {
		err = errors.New("Locking order table timed out")
		return
	}

	err = body()

	// we only unlock table if operation was successful
	if err == nil {
		err = ReleaseAnSetNxLockForTable(rc, "order")
		if err != nil {
			err = errors.Wrapf(err, "Failed to unlock orders table")
		}
	}

	return
}
