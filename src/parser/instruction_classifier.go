package parser

import (
	"github.com/go-solana-parse/src/config"
	"github.com/go-solana-parse/src/model"
)

// InstructionClassifier 指令分类器
type InstructionClassifier struct {
	adapter *TransactionAdapter
}

// NewInstructionClassifier 创建新的指令分类器
func NewInstructionClassifier(adapter *TransactionAdapter) *InstructionClassifier {
	return &InstructionClassifier{
		adapter: adapter,
	}
}

// GetAllProgramIDs 获取交易中的所有程序 ID
func (ic *InstructionClassifier) GetAllProgramIDs() []string {
	programIDs := make(map[string]bool)

	// 从主要指令获取程序 ID
	for _, instruction := range ic.adapter.GetInstructions() {
		programID := ic.adapter.GetProgramIdFromInstruction(instruction)
		if programID != "" {
			programIDs[programID] = true
		}
	}

	// 从内部指令获取程序 ID
	for _, innerInstructionGroup := range ic.adapter.GetInnerInstructions() {
		for _, innerInstruction := range innerInstructionGroup.Instructions {
			programID := ic.adapter.GetProgramIdFromInstruction(innerInstruction)
			if programID != "" {
				programIDs[programID] = true
			}
		}
	}

	// 转换为切片
	var result []string
	for programID := range programIDs {
		result = append(result, programID)
	}

	return result
}

// GetInstructions 获取特定程序 ID 的指令
func (ic *InstructionClassifier) GetInstructions(programID string) []model.ClassifiedInstruction {
	var classified []model.ClassifiedInstruction

	// 处理主要指令
	for i, instruction := range ic.adapter.GetInstructions() {
		if ic.adapter.GetProgramIdFromInstruction(instruction) == programID {
			classified = append(classified, model.ClassifiedInstruction{
				ProgramID:   programID,
				Instruction: instruction,
				OuterIndex:  i,
			})
		}
	}

	// 处理内部指令
	for _, innerInstructionGroup := range ic.adapter.GetInnerInstructions() {
		var innerInstructions []model.TransactionInstruction
		for _, innerInstruction := range innerInstructionGroup.Instructions {
			if ic.adapter.GetProgramIdFromInstruction(innerInstruction) == programID {
				innerInstructions = append(innerInstructions, innerInstruction)
			}
		}
		if len(innerInstructions) > 0 {
			classified = append(classified, model.ClassifiedInstruction{
				ProgramID:   programID,
				Instruction: innerInstructions,
				OuterIndex:  innerInstructionGroup.Index,
			})
		}
	}

	return classified
}

// GetDexInstructions 获取 DEX 相关的指令
func (ic *InstructionClassifier) GetDexInstructions() map[string][]model.ClassifiedInstruction {
	dexInstructions := make(map[string][]model.ClassifiedInstruction)

	allProgramIDs := ic.GetAllProgramIDs()
	for _, programID := range allProgramIDs {
		if _, exists := config.GetDexProgramByID(programID); exists {
			instructions := ic.GetInstructions(programID)
			if len(instructions) > 0 {
				dexInstructions[programID] = instructions
			}
		}
	}

	return dexInstructions
}

// HasDexInstructions 检查是否包含 DEX 指令
func (ic *InstructionClassifier) HasDexInstructions() bool {
	allProgramIDs := ic.GetAllProgramIDs()
	for _, programID := range allProgramIDs {
		if _, exists := config.GetDexProgramByID(programID); exists {
			return true
		}
	}
	return false
}

// GetJupiterInstructions 获取 Jupiter 相关的指令
func (ic *InstructionClassifier) GetJupiterInstructions() []model.ClassifiedInstruction {
	var jupiterInstructions []model.ClassifiedInstruction

	allProgramIDs := ic.GetAllProgramIDs()
	for _, programID := range allProgramIDs {
		if config.IsJupiterFamily(programID) {
			instructions := ic.GetInstructions(programID)
			jupiterInstructions = append(jupiterInstructions, instructions...)
		}
	}

	return jupiterInstructions
}
