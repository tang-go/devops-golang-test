# DevOps开发 golang 测试

1. 72小时内完成
2. fork本仓库
3. 通过kubebuilder或者手动创建golang的项目；完成自定义CRD MyStatefulSet核心功能的开发（功能同kubernetes StatefulSet），尽可能的完善！
4. 同时还需要实现ValidatingAdmissionWebhook，为MyStatefulSet提供准入验证
5. 要求不能直接引用kubernetes StatefulSet模块源码
6. 单元测试覆盖率必须大于80%，提交PR的时候，请附带上单元测试覆盖率截图或记录
7. 必须包含controller、AdmissionWebhook部署需要helm chart
8. makefile包含完整的编译流程（controller、AdmissionWebhook镜像的编译、helm chart编译、单元测试）
9. 完成以后通过pull request 提交，并备注面试姓名+联系方式，然后即时联系HR以免超时；

谢谢合作


# 完成度说明
1. 按照之前的要求来写的 MyRep1icaSet(功能同 kubernetes ReplicaSet)
2. 照猫画虎，实现了基本功能：修改 replicas 的值后能上报事件并自动拉起或删除 Pod，但是通过 kubectl delete pod 后无法拉起新的 Pod（事件没上报）
3. helm 部署后的权限问题暂未解决
4. 未做 单元测试

# 本人开发能力说明
1. 了解 golang 基本语法，只用过 golang 写过一些简单的脚本，未参与过大型项目的开发
2. 日常主要是做交付/运维工作，会看一些源码来辅助排错
3. 有强烈的意愿做 k8s 相关的开发工作
