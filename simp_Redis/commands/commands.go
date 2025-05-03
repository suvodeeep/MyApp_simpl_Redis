// Through the Package commands, we're implementing "Redis-like" command handlers
package commands

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"simp_Redis/store"
)

type Command func(s *store.Store, args []string) string

var commandMap = map[string]Command{
	"get":    getCommand,
	"set":    setCommand,
	"setnx":  setNXCommand,
	"del":    deleteCommand,
	"keys":   keysCommand,
	"ping":   pingCommand,
	"exists": existsCommand,
}

// Process handles a command string and executes the appropriate handler
func Process(s *store.Store, cmdLine string) string {

	cmdLine = strings.TrimSpace(cmdLine)

	parts := strings.Fields(cmdLine)
	if len(parts) == 0 {
		return "-ERR empty command"
	}

	cmdName := strings.ToLower(parts[0])
	args := parts[1:]

	cmd, exists := commandMap[cmdName]
	if !exists {
		return fmt.Sprintf("-ERR unknown command '%s'", cmdName)
	}

	return cmd(s, args)
}

// getCommand handles the GET command
func getCommand(s *store.Store, args []string) string {
	if len(args) != 1 {
		return "-ERR wrong number of arguments for 'get' command"
	}

	value, err := s.Get(args[0])
	if err == store.ErrKeyNotFound {
		return "$-1" // The Redis protocol for nil value
	}

	return fmt.Sprintf("$%d\r\n%s", len(value), value)
}

// setCommand handles the SET command

func setCommand(s *store.Store, args []string) string {
	if len(args) < 2 {
		return "-ERR wrong number of arguments for 'set' command"
	}

	key := args[0]
	value := args[1]
	var expiration *time.Time

	if len(args) > 2 && strings.ToLower(args[2]) == "ex" && len(args) > 3 {
		seconds, err := strconv.Atoi(args[3])
		if err != nil {
			return "-ERR value is not an integer or out of range"
		}

		if seconds <= 0 {
			return "-ERR invalid expire time in 'set' command"
		}

		exp := time.Now().Add(time.Duration(seconds) * time.Second)
		expiration = &exp
	}

	s.Set(key, value, expiration)
	return "+OK"
}

// setNXCommand  (Set if Not Exists) handles the SETNX command
func setNXCommand(s *store.Store, args []string) string {
	if len(args) < 2 {
		return "-ERR wrong number of arguments for 'setnx' command"
	}

	key := args[0]
	value := args[1]
	var expiration *time.Time

	if len(args) > 2 && strings.ToLower(args[2]) == "ex" && len(args) > 3 {
		seconds, err := strconv.Atoi(args[3])
		if err != nil {
			return "-ERR value is not an integer or out of range"
		}

		if seconds <= 0 {
			return "-ERR invalid expire time in 'setnx' command"
		}

		exp := time.Now().Add(time.Duration(seconds) * time.Second)
		expiration = &exp
	}

	// Try to set the value
	err := s.SetNX(key, value, expiration)
	if err == store.ErrKeyExists {
		return ":0" // Redis returns 0, if the key already exists
	}
	return ":1" // Redis returns 1, if the key was set
}

// deleteCommand here, handles the DEL command
func deleteCommand(s *store.Store, args []string) string {
	if len(args) < 1 {
		return "-ERR wrong number of arguments for 'del' command"
	}

	count := 0
	for _, key := range args {
		if s.Delete(key) {
			count++
		}
	}

	return fmt.Sprintf(":%d", count)
}

func keysCommand(s *store.Store, args []string) string {

	keys := s.Keys()

	result := fmt.Sprintf("*%d", len(keys))
	for _, key := range keys {
		result += fmt.Sprintf("\r\n$%d\r\n%s", len(key), key)
	}

	return result
}

// pingCommand implements the PING command If an argument is provided, it should echo it back
func pingCommand(s *store.Store, args []string) string {
	if len(args) == 0 {
		return "+PONG"
	}

	return fmt.Sprintf("$%d\r\n%s", len(args[0]), args[0])
}

func existsCommand(s *store.Store, args []string) string {
	if len(args) < 1 {
		return "-ERR wrong number of arguments for 'exists' command"
	}

	count := 0
	for _, key := range args {
		_, err := s.Get(key)
		if err == nil {
			count++
		}
	}

	return fmt.Sprintf(":%d", count)
}
