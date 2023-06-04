# localq 

###### project description

### Title 

Write the read me here...

### MasterQ(previously Queue)
  - organized and houses one or more TaskQueue
  - this is the root directory

### TaskQueue
  - represents the queue of a specific type of Task
  - this is the directory housing all instances of
    a type of task

### TaskInstance (previously TaskQueue)
  - represents an instance of a specific type of Task 
  - this is the representation of the task details on
    disk - the unique id, the argument data, access and errors.

### TaskExecutor (previously Task)
  - interface for the code that actually runs the instance
    of a task. This is what gets registered by the QueueManager.
  - Implementations of TaskExecutor is what is 
    registered with the LocalQ and executes the code