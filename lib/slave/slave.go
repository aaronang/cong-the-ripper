package slave

import (
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/aaronang/cong-the-ripper/lib"
)

var slaveInstance Slave

type Slave struct {
	port            string
	masterIp        string
	masterPort      string
	heartbeat       lib.Heartbeat
	successChan     chan CrackerSuccess
	failChan        chan CrackerFail
	addTaskChan     chan lib.Task
	heartbeatTicker *time.Ticker
}

func Init(instanceId, port, masterIp, masterPort string) *Slave {
	heartbeat := lib.Heartbeat{
		SlaveId: instanceId,
	}

	slaveInstance = Slave{
		port:            port,
		masterIp:        masterIp,
		masterPort:      masterPort,
		heartbeat:       heartbeat,
		successChan:     make(chan CrackerSuccess),
		failChan:        make(chan CrackerFail),
		addTaskChan:     make(chan lib.Task),
		heartbeatTicker: nil,
	}
	return &slaveInstance
}

func (s *Slave) Run() {
	log.Println("Running slave on port", s.port)

	go func() {
		http.HandleFunc(lib.TasksCreatePath, taskHandler)
		err := http.ListenAndServe(":"+s.port, nil)
		if err != nil {
			log.Panicln("[Main Loop] listener failed", err)
		}
	}()

	s.heartbeatTicker = time.NewTicker(lib.HeartbeatInterval)
	for {
		select {
		case <-s.heartbeatTicker.C:
			s.sendHeartbeat()

		case task := <-s.addTaskChan:
			log.Println("[Main Loop]", "Add task:", task)
			s.addTask(task)
			go Execute(task, s.successChan, s.failChan)

		case msg := <-s.successChan:
			log.Println("[Main Loop]", "SuccessChan message:", msg)
			s.passwordFound(msg.taskID, msg.password)

		case msg := <-s.failChan:
			log.Println("[Main Loop]", "FailChan message:", msg)
			s.passwordNotFound(msg.taskID)
		}
	}
}

func taskHandler(w http.ResponseWriter, r *http.Request) {
	var t lib.Task
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&t); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	slaveInstance.addTaskChan <- t
}

func (s *Slave) addTask(task lib.Task) {
	taskStatus := lib.TaskStatus{
		Id:       task.ID,
		JobId:    task.JobID,
		Status:   lib.Running,
		Progress: task.Start,
	}
	s.heartbeat.TaskStatus = append(s.heartbeat.TaskStatus, taskStatus)
}

func (s *Slave) passwordFound(id int, password string) {
	log.Println("[ Task", id, "]", "Found password:", password)
	ts := s.taskStatusWithId(id)
	if ts != nil {
		ts.Status = lib.PasswordFound
		ts.Password = password
	} else {
		log.Println("[ERROR]", "Id not found in Taskstatus")
	}
}

func (s *Slave) passwordNotFound(id int) {
	log.Println("[ Task", id, "]", "Password not found")
	ts := s.taskStatusWithId(id)
	if ts != nil {
		ts.Status = lib.PasswordNotFound
	} else {
		log.Println("ERROR:", "Id not found in Taskstatus")
	}
}

func (s *Slave) taskStatusWithId(id int) *lib.TaskStatus {
	for i, ts := range s.heartbeat.TaskStatus {
		if ts.Id == id {
			return &s.heartbeat.TaskStatus[i]
		}
	}
	return nil
}
