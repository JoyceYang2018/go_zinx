package core

import (
	"fmt"
	"testing"
)

func TestNewAOIManager(t *testing.T) {
	// 初始化AOIManager
	aoiMgr:=NewAOIManager(0,250,5,0,250,5)

	// 打印AOIManager
	fmt.Println(aoiMgr)
}

func TestAOIManagerSurroundGridsByGrid(t *testing.T) {
	// 初始化AOIManager
	aoiMgr:=NewAOIManager(0,250,5,0,250,5)

	for gid,_ := range aoiMgr.grids{
		grids := aoiMgr.GetSurroundGridsByGrid(gid)
		fmt.Println("gid:", gid, "grids len:", len(grids))
		gIDs := make([]int, 0, len(grids))
		for _, grid:= range grids{
			gIDs = append(gIDs, grid.GID)
		}
		fmt.Println("arrounding grid IDs are ", gIDs)
	}
}