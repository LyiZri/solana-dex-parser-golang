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
	ports:   make([]int, 0, 31), // 8000-8030 共31个端口
	current: 0,
}

// 初始化负载均衡器
func init() {
	// 初始化端口范围 8000-8030
	for port := 8000; port <= 8030; port++ {
		globalLoadBalancer.ports = append(globalLoadBalancer.ports, port)
	}
}

// 获取下一个可用端口（轮询算法）
func (lb *LoadBalancer) getNextPort() int {
	lb.mutex.Lock()
	defer lb.mutex.Unlock()

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
