package vctui

import (
	"context"
	"fmt"

	"github.com/gdamore/tcell"
	"github.com/rivo/tview"
	"github.com/vmware/govmomi/object"
	"github.com/vmware/govmomi/vim25/mo"
)

//MainUI starts up the katbox User Interface
func MainUI(v []*object.VirtualMachine) error {
	// Check for a nil pointer
	if v == nil {
		return fmt.Errorf("No VMs")
	}

	root := buildTree(v)

	tree := tview.NewTreeView().
		SetRoot(root).
		SetCurrentNode(root)
	application := tview.NewApplication()

	// If a directory was selected, open it.
	tree.SetSelectedFunc(func(node *tview.TreeNode) {
		reference := node.GetReference()
		if reference == nil {
			return // Selecting the root node does nothing.
		}
		children := node.GetChildren()
		// If it has children then flip the expanded state, if it's the final child we will action it
		if len(children) != 0 {
			node.SetExpanded(!node.IsExpanded())
		} else {
			// TODO - Open the action menu on the specific article
		}
	})

	tree.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Key() {
		case tcell.KeyCtrlF:
			// Stop the existing UI

			var subset []*object.VirtualMachine
			application.Suspend(func() { subset = SearchUI(v) })
			uiBugFix()
			// Get new tree
			newRoot := buildTree(subset)
			root.ClearChildren()
			root.SetChildren(newRoot.GetChildren())
		default:
			return event
		}
		return nil
	})

	if err := application.SetRoot(tree, true).Run(); err != nil {
		panic(err)
	}

	fmt.Printf("Suppose I should save some changes?\n")

	return nil
}

// This function will take the full article set and build a tree from any search parameters
func buildTree(v []*object.VirtualMachine) *tview.TreeNode {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	// Begin the UI Tree
	rootDir := "VMware vCenter"
	root := tview.NewTreeNode(rootDir).
		SetColor(tcell.ColorRed)

		// reference is used to label the type of tree Node
	var reference string

	// Add Github articles to the tree
	reference = "Virtual Machines"
	vmNode := tview.NewTreeNode("VMs").SetReference(reference).SetSelectable(true)
	vmNode.SetColor(tcell.ColorGreen)

	// Add Github articles to the tree
	reference = "Templates"
	templateNode := tview.NewTreeNode("Templates").SetReference(reference).SetSelectable(true)
	templateNode.SetColor(tcell.ColorGreen)

	for x := range v {
		vmChildNode := tview.NewTreeNode(v[x].Name()).SetReference(reference).SetSelectable(true).SetExpanded(false)
		vmChildNode.SetColor(tcell.ColorBlue)

		// Retrieve the managed object (using the summary string)
		var o mo.VirtualMachine

		err := v[x].Properties(ctx, v[x].Reference(), []string{"summary"}, &o)
		if err != nil {
			break
		}

		vmDetails := buildDetails(ctx, v[x], o)
		vmChildNode.AddChild(vmDetails)

		if o.Summary.Config.Template == true {
			templateNode.AddChild(vmChildNode)
		} else {
			vmNode.AddChild(vmChildNode)
		}
	}

	root.AddChild(vmNode)
	root.AddChild(templateNode)

	return root
}

func buildDetails(ctx context.Context, vm *object.VirtualMachine, vmo mo.VirtualMachine) *tview.TreeNode {

	// Add Details subtree information
	vmDetails := tview.NewTreeNode("Details").SetReference("Details").SetSelectable(true)

	vmDetail := tview.NewTreeNode(fmt.Sprintf("CPUs: %d", vmo.Summary.Config.NumCpu)).SetReference("Cpu").SetSelectable(true)
	vmDetails.AddChild(vmDetail)

	vmDetail = tview.NewTreeNode(fmt.Sprintf("Memory: %d", vmo.Summary.Config.MemorySizeMB)).SetReference("memory").SetSelectable(true)
	vmDetails.AddChild(vmDetail)

	vmDetail = tview.NewTreeNode(fmt.Sprintf("VMware Tools: %s", vmo.Summary.Guest.ToolsRunningStatus)).SetReference("toolsStatus").SetSelectable(true)
	vmDetails.AddChild(vmDetail)

	vmDetail = tview.NewTreeNode(fmt.Sprintf("VM IP Address: %s", vmo.Summary.Guest.IpAddress)).SetReference("toolsStatus").SetSelectable(true)
	vmDetails.AddChild(vmDetail)

	devices, _ := vm.Device(ctx)

	//for _, device := range devices {

	vmDetail = tview.NewTreeNode(fmt.Sprintf("MAC ADDRESS: %s", devices.PrimaryMacAddress())).SetReference("toolsStatus").SetSelectable(true)
	vmDetails.AddChild(vmDetail)
	//		device.GetVirtualDevice().
	//}
	// if v.Config != nil {

	// 	if len(v.Config.Hardware.Device) != 0 {
	// 		for x := range v.Config.Hardware.Device {
	// 			if nic, ok := v.Config.Hardware.Device[x].(types.BaseVirtualEthernetCard); ok {
	// 				mac := nic.GetVirtualEthernetCard().MacAddress
	// 				if mac != "" {
	// 					vmDetail = tview.NewTreeNode(fmt.Sprintf("MAC ADDRESS: %s", mac)).SetReference("toolsStatus").SetSelectable(true)
	// 					vmDetails.AddChild(vmDetail)
	// 					//return false
	// 				}
	// 				//				macs[mac] = nil
	// 				//				eths[devices.Name(d)] = mac
	// 			}
	// 		}
	// 	}
	// } else {
	// }
	return vmDetails
}
