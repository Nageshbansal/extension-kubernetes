// SPDX-License-Identifier: MIT
// SPDX-FileCopyrightText: 2022 Steadybit GmbH

package extstatefulset

const (
	StatefulSetTargetType    = "com.steadybit.extension_kubernetes.kubernetes-statefulset"
	statefulSetIcon          = "data:image/svg+xml;base64,PHN2ZyB3aWR0aD0iMjQiIGhlaWdodD0iMjQiIHZpZXdCb3g9IjAgMCAyNCAyNCIgZmlsbD0ibm9uZSIgeG1sbnM9Imh0dHA6Ly93d3cudzMub3JnLzIwMDAvc3ZnIj4KPHBhdGggZmlsbC1ydWxlPSJldmVub2RkIiBjbGlwLXJ1bGU9ImV2ZW5vZGQiIGQ9Ik0xOS41NjI1IDEzLjI1VjExLjc1SDIxQzIxLjI3NjEgMTEuNzUgMjEuNSAxMS41MjYxIDIxLjUgMTEuMjVWOS4zMTI1SDIzVjExLjI1QzIzIDEyLjM1NDYgMjIuMTA0NiAxMy4yNSAyMSAxMy4yNUgxOS41NjI1Wk0yMyA1LjQzNzVIMjEuNVYzLjVDMjEuNSAzLjIyMzg2IDIxLjI3NjEgMyAyMSAzSDE5LjU2MjVWMS41SDIxQzIyLjEwNDYgMS41IDIzIDIuMzk1NDMgMjMgMy41VjUuNDM3NVpNMTYuNjg3NSAxLjVWM0gxMy44MTI1VjEuNUgxNi42ODc1Wk0xMC45Mzc1IDEuNVYzSDkuNUM5LjIyMzg2IDMgOSAzLjIyMzg2IDkgMy41VjUuNDM3NUg3LjVWMy41QzcuNSAyLjM5NTQzIDguMzk1NDMgMS41IDkuNSAxLjVIMTAuOTM3NVpNMTcgMTEuMzYxMlY5LjkwMDE3QzE3IDguNjEyNjIgMTMuNjQwNiA3LjU2MjUgOS41IDcuNTYyNUM1LjM1OTM2IDcuNTYyNSAyIDguNjEyNjIgMiA5LjkwMDE3VjExLjM2MTJDMiAxMi42NDg3IDUuMzU5MzYgMTMuNjk4OSA5LjUgMTMuNjk4OUMxMy42NDA2IDEzLjY5ODkgMTcgMTIuNjQ4NyAxNyAxMS4zNjEyWk0xNyAxOC45NzQ4VjE0LjQzNzVDMTUuMzg4NiAxNS40OTY4IDEyLjQzOTUgMTUuOTg5OSA5LjUgMTUuOTg5OUM2LjU2MDU0IDE1Ljk4OTkgMy42MTEzMyAxNS40OTY4IDIgMTQuNDM3NVYxOC45NzQ4QzIgMjAuMjYyNCA1LjM1OTM2IDIxLjMxMjUgOS41IDIxLjMxMjVDMTMuNjQwNiAyMS4zMTI1IDE3IDIwLjI2MjQgMTcgMTguOTc0OFoiIGZpbGw9IiMxRDI2MzIiLz4KPC9zdmc+Cg=="
	ScaleStatefulSetActionId = "com.steadybit.extension_kubernetes.scale_statefulset"
	scaleStatefulSetIcon     = "data:image/svg+xml;base64,PHN2ZyB3aWR0aD0iMjQiIGhlaWdodD0iMjQiIHZpZXdCb3g9IjAgMCAyNCAyNCIgZmlsbD0ibm9uZSIgeG1sbnM9Imh0dHA6Ly93d3cudzMub3JnLzIwMDAvc3ZnIj4KPHBhdGggZmlsbC1ydWxlPSJldmVub2RkIiBjbGlwLXJ1bGU9ImV2ZW5vZGQiIGQ9Ik0xOS41NjI1IDEzLjI1VjExLjc1SDIxQzIxLjI3NjEgMTEuNzUgMjEuNSAxMS41MjYxIDIxLjUgMTEuMjVWOS4zMTI1SDIzVjExLjI1QzIzIDEyLjM1NDYgMjIuMTA0NiAxMy4yNSAyMSAxMy4yNUgxOS41NjI1Wk0yMyA1LjQzNzVIMjEuNVYzLjVDMjEuNSAzLjIyMzg2IDIxLjI3NjEgMyAyMSAzSDE5LjU2MjVWMS41SDIxQzIyLjEwNDYgMS41IDIzIDIuMzk1NDMgMjMgMy41VjUuNDM3NVpNMTYuNjg3NSAxLjVWM0gxMy44MTI1VjEuNUgxNi42ODc1Wk0xMC45Mzc1IDEuNVYzSDkuNUM5LjIyMzg2IDMgOSAzLjIyMzg2IDkgMy41VjUuNDM3NUg3LjVWMy41QzcuNSAyLjM5NTQzIDguMzk1NDMgMS41IDkuNSAxLjVIMTAuOTM3NVpNMTcgMTEuMzYxMlY5LjkwMDE3QzE3IDguNjEyNjIgMTMuNjQwNiA3LjU2MjUgOS41IDcuNTYyNUM1LjM1OTM2IDcuNTYyNSAyIDguNjEyNjIgMiA5LjkwMDE3VjExLjM2MTJDMiAxMi42NDg3IDUuMzU5MzYgMTMuNjk4OSA5LjUgMTMuNjk4OUMxMy42NDA2IDEzLjY5ODkgMTcgMTIuNjQ4NyAxNyAxMS4zNjEyWk0xNyAxOC45NzQ4VjE0LjQzNzVDMTUuMzg4NiAxNS40OTY4IDEyLjQzOTUgMTUuOTg5OSA5LjUgMTUuOTg5OUM2LjU2MDU0IDE1Ljk4OTkgMy42MTEzMyAxNS40OTY4IDIgMTQuNDM3NVYxOC45NzQ4QzIgMjAuMjYyNCA1LjM1OTM2IDIxLjMxMjUgOS41IDIxLjMxMjVDMTMuNjQwNiAyMS4zMTI1IDE3IDIwLjI2MjQgMTcgMTguOTc0OFoiIGZpbGw9IiMxRDI2MzIiLz4KPC9zdmc+Cg=="
)
