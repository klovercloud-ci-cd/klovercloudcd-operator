package utility

import (
	"bytes"
	"fmt"
	v1 "k8s.io/api/core/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/remotecommand"
)

func ExecuteRemoteCommand(restCfg *rest.Config, podName, namespace string, container string, commands []string) (data string, errData string, err error) {
	var commandArr []string
	commandArr = append(commandArr, "/bin/sh")
	commandArr = append(commandArr, "-c")
	commandArr = append(commandArr, commands...)
	coreClient, err := kubernetes.NewForConfig(restCfg)
	if err != nil {
		return "", "", err
	}
	buf := &bytes.Buffer{}
	errBuf := &bytes.Buffer{}
	request := coreClient.CoreV1().RESTClient().
		Post().
		Namespace(namespace).
		Resource("pods").
		Name(podName).
		SubResource("exec").
		Param("container", container).
		VersionedParams(&v1.PodExecOptions{
			Command: commandArr,
			Stdin:   false,
			Stdout:  true,
			Stderr:  true,
			TTY:     true,
		}, scheme.ParameterCodec)
	exec, err := remotecommand.NewSPDYExecutor(restCfg, "POST", request.URL())
	err = exec.Stream(remotecommand.StreamOptions{
		Stdout: buf,
		Stderr: errBuf,
	})
	if err != nil {
		return "", "", fmt.Errorf("%w Failed executing command %s on %v/%v", err, commands, namespace, podName)
	}

	return buf.String(), errBuf.String(), nil
}
