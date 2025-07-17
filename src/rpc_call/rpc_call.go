package rpccall

import (
	"fmt"
	"sync"

	"github.com/go-solana-parse/src/model"
	"github.com/go-solana-parse/src/util"
)

// 负载均衡器结构
type LoadBalancer struct {
	ports   []int
	current int
	mutex   sync.Mutex
}

// 全局负载均衡器实例
var globalLoadBalancer = &LoadBalancer{
	ports:   make([]int, 0, 10), // 预留容量
	current: 0,
}

// 初始化负载均衡器 - 默认端口范围
func init() {
	// 默认初始化为8000-8009（10个端口）
	setPortsFromStart(8000, 10)
}

// 设置端口范围（从指定端口开始，使用6个端口）
func SetPortRange(startPort int) {
	globalLoadBalancer.mutex.Lock()
	defer globalLoadBalancer.mutex.Unlock()

	setPortsFromStart(startPort, 6) // 每个进程分配6个端口
	globalLoadBalancer.current = 0  // 重置计数器

	fmt.Printf("🌐 端口范围已设置: %d-%d\n", startPort, startPort+5)
}

// 内部函数：设置端口列表
func setPortsFromStart(startPort, count int) {
	globalLoadBalancer.ports = make([]int, 0, count)
	for i := 0; i < count; i++ {
		globalLoadBalancer.ports = append(globalLoadBalancer.ports, startPort+i)
	}
}

// 获取下一个可用端口（轮询算法）
func (lb *LoadBalancer) getNextPort() int {
	lb.mutex.Lock()
	defer lb.mutex.Unlock()

	if len(lb.ports) == 0 {
		return 8000 // 安全回退
	}

	port := lb.ports[lb.current]
	lb.current = (lb.current + 1) % len(lb.ports)
	return port
}

// 构建URL
func (lb *LoadBalancer) getNextURL() string {
	port := lb.getNextPort()
	return fmt.Sprintf("http://localhost:%d/api/parse-blockdata", port)
}

func SendParseDataToDeno(blockNum string, blockData model.Block) error {

	req := model.ParseBlockDataDenoReq{
		BlockNum:  blockNum,
		BlockData: blockData,
	}

	// 使用负载均衡获取URL
	url := globalLoadBalancer.getNextURL()
	body, err := util.PostReq(url, req)
	if err != nil {
		return err
	}

	fmt.Println(string(body))
	return nil
}

func SendMultipleParseDataToDeno(data []model.ParseBlockDataDenoReq) error {

	// 使用负载均衡获取URL
	url := globalLoadBalancer.getNextURL()
	body, err := util.PostReq(url, data)
	if err != nil {
		return err
	}

	fmt.Printf("Request sent to %s\n", url)
	fmt.Println(string(body))
	return nil
}
