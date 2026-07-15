package brokerdto

type FixationCustomerRequest struct {
	CustomerID string `json:"customer_id" validate:"required,uuid4"` //мешают ли для мапингу приватные типы, если да - почему?
	ManagerID  string `json:"manager_id" validate:"required,uuid4"`  //мы должны ManagerID вытащить из принципала, а CustomerID из query параметра? как вообще в body попадают данные? допустим когда есть ui какой-то и с ui отправляют, как в случае моего проекта
}

//type ConnectCustomerResponse struct { //верно понял что мы ничего не должны возвращать кроме ошибки или http 201? флоу такой - свободный клиент(кастомер), ты его закрепляешь за собой, видишь свои фио на кастомере что он за тобой. Мне надо возвращать фио при удачном закреплении или нет? как принято делать?
//	ManagerLastName   string `json:"manager_last_name"`
//	ManagerFirstName  string `json:"manager_first_name"`
//	ManagerMiddleName string `json:"manager_middle_name"`
//}
