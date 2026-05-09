package common_battle

import "g7/common/protos/pb"

// Vector3D X前后 Y左右 Z上下
type Vector3D struct {
	X float64
	Y float64
	Z float64
}

func (v3d *Vector3D) GetXY2D() (x, y int32) {
	return v3d.X, v3d.Y
}

func (v3d *Vector3D) ToProto() *pb.Action_Move {
	return &pb.Action_Move{X: v3d.X, Y: v3d.Y, Z: v3d.Z}
}

func (v3d *Vector3D) Add(add Vector3D) Vector3D {
	return *v3d
}

func (v3d *Vector3D) Mul(f float64) Vector3D {
	return *v3d
}
