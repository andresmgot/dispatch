///////////////////////////////////////////////////////////////////////
// Copyright (c) 2017 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0
///////////////////////////////////////////////////////////////////////

package kubeless

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"reflect"
	"strconv"
	"strings"
	"time"

	"github.com/kubeless/kubeless/pkg/apis/kubeless/v1beta1"
	"github.com/kubeless/kubeless/pkg/client/clientset/versioned"
	"github.com/vmware/dispatch/pkg/functions"
	k8sErrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

type kubelessDriver struct {
	config      *Config
	kubelessCli *versioned.Clientset
	k8sClient   *kubernetes.Clientset
	fnNs        string
}

// Config contains the OpenWhisk configuration
type Config struct {
	K8sConfig     string
	FuncNamespace string
}

func New(config *Config) (functions.FaaSDriver, error) {
	k8sConf, err := kubeClientConfig(config.K8sConfig)
	if err != nil {
		log.Fatalf("error configuring k8s API client: %v", err)
	}
	kubelessCli := versioned.NewForConfigOrDie(k8sConf)
	k8sClient := kubernetes.NewForConfigOrDie(k8sConf)

	fnNs := config.FuncNamespace
	if fnNs == "" {
		fnNs = "default"
	}

	return &kubelessDriver{config, kubelessCli, k8sClient, fnNs}, nil
}

func kubeClientConfig(kubeConfPath string) (*rest.Config, error) {
	if kubeConfPath != "" {
		return clientcmd.BuildConfigFromFlags("", kubeConfPath)
	}
	return rest.InClusterConfig()
}

func (d *kubelessDriver) Create(f *functions.Function, exec *functions.Exec) error {
	h := sha256.New()
	_, err := h.Write([]byte(f.Code))
	if err != nil {
		return fmt.Errorf("Unable to obtain dependencies checksum: %v", err)
	}
	checksum := hex.EncodeToString(h.Sum(nil))
	kf := v1beta1.Function{
		ObjectMeta: metav1.ObjectMeta{
			Name:      f.Name,
			Namespace: d.fnNs,
			Labels: map[string]string{
				"ID": f.FaasID,
			},
		},
		Spec: v1beta1.FunctionSpec{
			Handler:             "handler." + f.Main,
			Function:            f.Code,
			Runtime:             f.Runtime,
			FunctionContentType: "text",
			Checksum:            fmt.Sprintf("sha256:%s", checksum),
		},
	}
	res, err := d.kubelessCli.KubelessV1beta1().Functions(d.fnNs).Create(&kf)
	if err != nil {
		return err
	}
	log.Printf("Function! %v", *res)
	return nil
}

func (d *kubelessDriver) Delete(f *functions.Function) error {
	err := d.kubelessCli.KubelessV1beta1().Functions(d.fnNs).Delete(f.Name, &metav1.DeleteOptions{})
	if err != nil && !k8sErrors.IsNotFound(err) {
		return err
	}
	return nil
}

func (d *kubelessDriver) getFuncName(id string) (string, error) {
	list, err := d.kubelessCli.KubelessV1beta1().Functions(d.fnNs).List(metav1.ListOptions{LabelSelector: fmt.Sprintf("ID=%s", id)})
	if err != nil {
		return "", err
	}
	if len(list.Items) != 1 {
		return "", fmt.Errorf("Unexpected amount of functions found %v", list.Items)
	}
	return list.Items[0].Name, nil
}

func getHTTPReq(clientset kubernetes.Interface, funcName, namespace, eventID, eventNamespace, method, body string) (*http.Request, error) {
	svc, err := clientset.CoreV1().Services(namespace).Get(funcName, metav1.GetOptions{})
	if err != nil {
		return nil, fmt.Errorf("Unable to find the service for function %s", funcName)
	}
	funcPort := strconv.Itoa(int(svc.Spec.Ports[0].Port))
	req, err := http.NewRequest(method, fmt.Sprintf("http://%s.%s.svc.cluster.local:%s", funcName, namespace, funcPort), strings.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("Unable to create request %v", err)
	}
	timestamp := time.Now().UTC()
	req.Header.Add("event-id", eventID)
	req.Header.Add("event-time", timestamp.String())
	req.Header.Add("event-namespace", eventNamespace)
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("event-type", "application/json")
	return req, nil
}

func (d *kubelessDriver) GetRunnable(e *functions.FunctionExecution) functions.Runnable {
	return func(ctx functions.Context, in interface{}) (interface{}, error) {
		req := &http.Request{}
		name, err := d.getFuncName(e.FaasID)
		if err != nil {
			return nil, err
		}
		log.Println("Executing %s", name)
		method := "GET"
		body := ""
		if reflect.ValueOf(in).Len() > 0 {
			inBytes, err := json.Marshal(in)
			if err != nil {
				return nil, err
			}
			method = "POST"
			body = string(inBytes)
		}
		req, err = getHTTPReq(d.k8sClient, name, d.fnNs, e.RunID, "dispatch.vmware.github.io", method, body)
		if err != nil {
			return nil, err
		}
		client := &http.Client{}
		resp, err := client.Do(req)
		if err != nil {
			return nil, err
		}
		defer resp.Body.Close()
		if resp.StatusCode != 200 {
			return nil, fmt.Errorf("Error: received error code %d: %s", resp.StatusCode, resp.Status)
		}
		res, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}
		var out interface{}
		if err := json.Unmarshal(res, &out); err != nil {
			return string(res), nil
		}
		return out, nil
	}
}
