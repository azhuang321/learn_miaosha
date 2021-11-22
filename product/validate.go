package main

import (
	"errors"
	"fmt"
	"imooc-product/product/common"
	"imooc-product/product/encrypt"
	"io/ioutil"
	"net/http"
	"strconv"
	"sync"
	"time"

	"encoding/json"
	"imooc-product/product/datamodels"
	"imooc-product/product/rabbitmq"
	"net/url"
)

var (
	// 设置集群地址，最好是内网IP，比较快
	hostArray = []string{"127.0.0.1", "127.0.0.1"}

	localHost = ""

	port = "8081"

	hashConsistent *common.Consistent

	accessControl = &AccessControl{
		sourcesArray: make(map[int]time.Time),
	}

	rabbitmqValidate *rabbitmq.RabbitMQ

	//数量控制接口服务器内网IP，或者getone的SLB内网IP
	GetOneIp = "127.0.0.1"

	GetOnePort = "8084"

	// 服务器间隔时间
	interval = 20

	blackList = &BlackList{
		listArray: make(map[int]bool),
	}
)

// 存放控制信息
type AccessControl struct {
	// 存放用户想要的信息
	sourcesArray map[int]time.Time
	sync.RWMutex
}

// 黑名单
type BlackList struct {
	listArray map[int]bool
	sync.RWMutex
}

// 获取黑名单
func (b *BlackList) GetBlackListByID(uid int) bool {
	b.RLock()
	defer b.RUnlock()
	return b.listArray[uid]
}

// 设置黑名单
func (b *BlackList) SetBlackListByID(uid int) bool {
	b.Lock()
	defer b.Unlock()
	b.listArray[uid] = true
	return true
}

// 根据用户uid获取记录
func (a *AccessControl) GetNewRecord(uid int) time.Time {
	a.RWMutex.RLock()
	defer a.RWMutex.RUnlock()
	return a.sourcesArray[uid]
}

// 根据用户uid设置记录
func (a *AccessControl) SetNewRecord(uid int) {
	a.RWMutex.Lock()
	defer a.RWMutex.Unlock()
	a.sourcesArray[uid] = time.Now()
}

// 分布式之间的数据共享
func (a *AccessControl) GetDistributedRight(req *http.Request) bool {
	// 获取用户uid
	uid, err := req.Cookie("uid")
	if err != nil {
		return false
	}
	// 根据用户uid，判断获取具体机器，采用一致性hash算法
	hostRequest, err := hashConsistent.Get(uid.Value)
	if err != nil {
		return false
	}
	// 判断是否为本机
	if hostRequest == localHost {
		// 是本机执行本机数据读取和校验
		return a.GetDataFromMap(uid.Value)
	} else {
		// 不是本机充当代理访问数据返回结果
		return GetDataFromOtherMap(hostRequest, req)
	}
}

//获取本机map，并且处理业务逻辑，返回的结果类型为bool类型（判断用户在限定的时间抢购的次数）
func (a *AccessControl) GetDataFromMap(uid string) (isOk bool) {
	uidInt, err := strconv.Atoi(uid)
	if err != nil {
		return false
	}

	// 判断用户是否在黑名单
	if blackList.GetBlackListByID(uidInt) {
		return false
	}

	// 获取记录
	dataRecord := a.GetNewRecord(uidInt)
	if dataRecord.IsZero() {
		if dataRecord.Add(time.Duration(interval) * time.Second).After(time.Now()) {
			return false
		}
	}
	a.SetNewRecord(uidInt)
	return true
}

//获取其它节点处理结果
func GetDataFromOtherMap(host string, request *http.Request) bool {
	hostUrl := "http://" + host + ":" + port + "/checkRight"
	response, body, err := GetCurl(hostUrl, request)
	if err != nil {
		return false
	}
	//判断状态
	if response.StatusCode == 200 {
		if string(body) == "true" {
			return true
		} else {
			return false
		}
	}
	return false
}

//模拟请求
func GetCurl(hostUrl string, request *http.Request) (response *http.Response, body []byte, err error) {
	//获取Uid
	uidPre, err := request.Cookie("uid")
	if err != nil {
		return
	}
	//获取sign
	uidSign, err := request.Cookie("sign")
	if err != nil {
		return
	}
	//模拟接口访问，
	client := &http.Client{}
	req, err := http.NewRequest("GET", hostUrl, nil)
	if err != nil {
		return
	}
	//手动指定，排查多余cookies
	cookieUid := &http.Cookie{Name: "uid", Value: uidPre.Value, Path: "/"}
	cookieSign := &http.Cookie{Name: "sign", Value: uidSign.Value, Path: "/"}
	//添加cookie到模拟的请求中
	req.AddCookie(cookieUid)
	req.AddCookie(cookieSign)
	//获取返回结果
	response, err = client.Do(req)
	defer response.Body.Close()
	if err != nil {
		return
	}
	body, err = ioutil.ReadAll(response.Body)
	return
}

// 验证权限
func CheckRight(w http.ResponseWriter, r *http.Request) {
	right := accessControl.GetDistributedRight(r)
	if !right {
		w.Write([]byte("false"))
		return
	}
	w.Write([]byte("true"))
	return
}

//执行正常业务逻辑
func Check(w http.ResponseWriter, r *http.Request) {
	//执行正常业务逻辑
	fmt.Println("执行check！")
	queryForm, err := url.ParseQuery(r.URL.RawQuery)
	if err != nil || len(queryForm["productID"]) <= 0 {
		w.Write([]byte("false"))
		return
	}
	productString := queryForm["productID"][0]
	fmt.Println(productString)
	//获取用户cookie
	userCookie, err := r.Cookie("uid")
	if err != nil {
		w.Write([]byte("false"))
		return
	}

	//1.分布式权限验证
	right := accessControl.GetDistributedRight(r)
	if right == false {
		w.Write([]byte("false"))
		return
	}
	//2.获取数量控制权限，防止秒杀出现超卖现象
	hostUrl := "http://" + GetOneIp + ":" + GetOnePort + "/getOne"
	responseValidate, validateBody, err := GetCurl(hostUrl, r)
	if err != nil {
		w.Write([]byte("false"))
		return
	}
	//判断数量控制接口请求状态
	if responseValidate.StatusCode == 200 {
		if string(validateBody) == "true" {
			//整合下单
			//1.获取商品ID
			productID, err := strconv.ParseInt(productString, 10, 64)
			if err != nil {
				w.Write([]byte("false"))
				return
			}
			//2.获取用户ID
			userID, err := strconv.ParseInt(userCookie.Value, 10, 64)
			if err != nil {
				w.Write([]byte("false"))
				return
			}
			//3.创建消息体
			message := datamodels.NewMessage(userID, productID)
			//类型转化
			byteMessage, err := json.Marshal(message)
			if err != nil {
				w.Write([]byte("false"))
				return
			}
			//4.生产消息
			err = rabbitmqValidate.PublishSimple(string(byteMessage))
			if err != nil {
				w.Write([]byte("false"))
				return
			}
			w.Write([]byte("true"))
			return
		}
	}
	w.Write([]byte("false"))
	return
}

//统一验证拦截器，每个接口都需要提前验证
func Auth(w http.ResponseWriter, r *http.Request) error {
	fmt.Println("执行验证")
	// 基于Cookie的权限验证
	err := CheckUserInfo(r)
	if err != nil {
		return err
	}
	return nil
}

// 身份校验函数
func CheckUserInfo(r *http.Request) error {
	uidCookie, err := r.Cookie("uid")
	if err != nil {
		return errors.New("用户UID Cookie获取失败！")
	}
	// 获取用户加密串
	signCookie, err := r.Cookie("sign")
	if err != nil {
		return errors.New("用户加密串 Cookie获取失败！")
	}
	// 对信息进行解密
	signByte, err := encrypt.DePwdCode(signCookie.Value)
	if err != nil {
		return errors.New("用户加密串已被篡改！")
	}
	fmt.Println("结果开始比对")
	fmt.Println("用户ID：", uidCookie.Value)
	fmt.Println("解密后用户ID", string(signByte))
	if CheckInfo(uidCookie.Value, string(signByte)) {
		return nil
	}
	return errors.New("身份校验失败！")
}

// 比对 获取到的用户加密串 和 解密后的信息 是否一致
func CheckInfo(checkStr string, signStr string) bool {
	if checkStr == signStr {
		return true
	}
	return false
}

func main() {
	// 负载均衡器设置
	// 采用一致性hash算法
	hashConsistent = common.NewConsistent()
	// 采用一致性hash算法，添加节点
	for _, v := range hostArray {
		hashConsistent.Add(v)
	}

	// 自动获取ip，方便部署
	localIp, err := common.GetIntranceIp()
	if err != nil {
		fmt.Println(err)
	}
	localHost = localIp
	fmt.Println(localHost)

	// 创建rabbitmq
	rabbitmqValidate = rabbitmq.NewRabbitMQSimple("imoocProduct")
	defer rabbitmqValidate.Destory()

	// 设置静态文件目录（前端限流）
	http.Handle("/html/", http.StripPrefix("/html/", http.FileServer(http.Dir("./fronted/web/htmlProductShow"))))
	// 设置资源目录
	http.Handle("/public/", http.StripPrefix("/public/", http.FileServer(http.Dir("./fronted/web/public"))))

	// 1. 过滤器
	filter := common.NewFilter()
	// 注册拦截器
	filter.RegisterFilterUri("/check", Auth)
	filter.RegisterFilterUri("/checkRight", Auth)
	// 2. 启动服务
	http.HandleFunc("/check", filter.Handle(Check))
	http.HandleFunc("/checkRight", filter.Handle(CheckRight))
	http.ListenAndServe(":8083", nil)
}
