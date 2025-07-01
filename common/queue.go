package common

import (
	"container/list"
	"sync"
)

type Queue struct {
	Mutex sync.Mutex // 동시 수정 방지
	V     *list.List
}

/*
*******************************************************************************************

	Function    : CreateQ
	Description : QUEUE 생성
	Argumet     :
	Return      :
	ETC         :

******************************************************************************************
*/
func (q *Queue) CreateQ() {
	q.Mutex.Lock()
	defer q.Mutex.Unlock()

	q.V = list.New()
}

/*
*******************************************************************************************

	Function    : PushQ
	Description : QUEUE PUSH
	Argumet     : 1. (interface{}) PUSH 정보
	Return      :
	ETC         :

******************************************************************************************
*/
func (q *Queue) PushQ(val interface{}) {
	q.Mutex.Lock()
	defer q.Mutex.Unlock()

	q.V.PushBack(val)
}

/*
*******************************************************************************************

	Function    : PushQ_noMutex
	Description : QUEUE PUSH (Mutex 무시)
	Argumet     : 1. (interface{}) PUSH 정보
	Return      :
	ETC         :

******************************************************************************************
*/
func (q *Queue) PushQ_noMutex(val interface{}) {

	q.V.PushBack(val)
}

/*
*******************************************************************************************

	Function    : PopQ
	Description : QUEUE POP
	Argumet     :
	Return      : 1. (interface{}) POP 정보
	ETC         :

******************************************************************************************
*/
func (q *Queue) PopQ() interface{} {
	q.Mutex.Lock()
	defer q.Mutex.Unlock()

	front := q.V.Front()
	if front == nil {
		return nil
	}

	return q.V.Remove(front)
}

/*
*******************************************************************************************

	Function    : PopQ_noMutex
	Description : QUEUE POP (Mutex 무시)
	Argumet     :
	Return      : 1. (interface{}) POP 정보
	ETC         :

******************************************************************************************
*/
func (q *Queue) PopQ_noMutex() interface{} {
	front := q.V.Front()
	if front == nil {
		return nil
	}

	return q.V.Remove(front)
}

/*
*******************************************************************************************

	Function    : Count
	Description : QUEUE 정보 개수 조회
	Argumet     :
	Return      : 1. (int) QUEUE 정보 개수
	ETC         :

******************************************************************************************
*/
func (q *Queue) Count() int {
	q.Mutex.Lock()
	defer q.Mutex.Unlock()

	return q.V.Len()
}

/*
*******************************************************************************************

	Function    : Clear
	Description : QUEUE 초기화
	Argumet     :
	Return      :
	ETC         :

******************************************************************************************
*/
func (q *Queue) Clear() {
	q.Mutex.Lock()
	defer q.Mutex.Unlock()
	q.V.Init()
}
