# nursecall
Agent of https://nursecall.run web service.

nursecall is command wrapper and agent for web service of nursecall.run.

## about nursecall.run

準備中

## Example

```
export NURSECALL_CALL_TOKEN="(your call token)"
nursecall /var/www/batch/backup.sh
```

### Options

#### Change Task Name

You can override the default task name sent to Nursecall.  
This is useful when the command itself is generic (e.g., `/var/www/batch/backup.sh`) but you want a more descriptive name in the dashboard.  

```
export NURSECALL_TASK_NAME="DB Backup"
```

#### Use heartbeat

You can enable **heartbeat mode** by setting an interval in seconds.  
When enabled, the `nursecall` command will periodically send a heartbeat signal to the dashboard while the process is running.  
The dashboard updates the "last heartbeat received at" timestamp, so you can confirm the job is still alive.  
If the process hangs or stops unexpectedly and no heartbeat is received, you can recognize it as a potential failure by checking the dashboard.  


```
export NURSECALL_HEARTBEAT_INTERVAL_SEC=30
```

## Requirements

- Go

## Installation

```
go install github.com/narita-takeru/nursecall/cmd/nursecall
```


