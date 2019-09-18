package selfhandler

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/jonyhy96/containerM/client"
	"github.com/jonyhy96/containerM/models"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/filters"
	nt "github.com/docker/docker/api/types/network"
	dc "github.com/docker/docker/client"
	"github.com/docker/go-connections/nat"
)

// GoHandler implements Handler interface
type GoHandler struct {
	logger *log.Logger
}

// NewGoHandler return a new goHandler
func NewGoHandler(logger *log.Logger) *GoHandler {
	return &GoHandler{
		logger: logger,
	}
}

// SetupRoute setup route
func (g *GoHandler) SetupRoute(mux *http.ServeMux, ec chan error) {
	mux.HandleFunc("/hooks", func(w http.ResponseWriter, r *http.Request) {
		g.logger.Printf("Receive request from: %s\n", r.RemoteAddr)
		go g.Handler(r, ec)
		select {
		case res := <-ec:
			g.logger.Printf(res.Error())
			var resp = models.Response{
				Status:  500,
				Message: res.Error(),
			}
			jsonB, _ := json.Marshal(resp)
			w.WriteHeader(500)
			w.Write(jsonB)
		case <-time.After(5 * time.Second):
			var resp = models.Response{
				Status:  200,
				Message: "success",
			}
			jsonB, _ := json.Marshal(resp)
			w.WriteHeader(200)
			w.Write(jsonB)
		}
	})
}

// Handler handles
func (g *GoHandler) Handler(r *http.Request, ec chan error) {
	defer func() {
		if r := recover(); r != nil {
			g.logger.Println("Recovered err:", r)
		}
	}()
	ps := r.URL.Query()
	errs := checkParams(ps)
	if len(errs) > 0 {
		for _, e := range errs {
			ec <- e
			return
		}
	}
	var image = ps["image"][0]
	cli := client.GetCli()
	ctx := context.Background()
	tempFilter := filters.NewArgs(filters.KeyValuePair{
		Key:   "ancestor",
		Value: image,
	})
	skContainer, err := cli.ContainerList(ctx, types.ContainerListOptions{
		All:     true,
		Filters: tempFilter})
	if err != nil {
		g.logger.Printf("err:%+v\n", err)
		ec <- err
		return
	}
	if len(skContainer) > 0 {
		for _, c := range skContainer {
			err := g.deleteOldContainer(ctx, cli, c)
			if err != nil {
				ec <- err
				return
			}
		}
	}
	err = g.pullNewImage(ctx, cli, image)
	if err != nil {
		ec <- err
		return
	}
	skNet, err := g.getSkNetwork(ctx, cli)
	if err != nil {
		ec <- err
		return
	}
	var env []string
	if val, ok := ps["env"]; ok {
		var temp = map[string]interface{}{}
		err = json.Unmarshal([]byte(val[0]), &temp)
		if err != nil {
			ec <- err
			return
		}
		for k, v := range temp {
			env = append(env, k+"="+v.(string))
		}
		fmt.Printf("env:%+v\n", temp)
	}
	ports := make(nat.PortSet, 1)
	if val, ok := ps["ports"]; ok {
		for _, v := range strings.Split(val[0], ",") {
			ports[nat.Port(v)] = struct{}{}
		}
		fmt.Printf("ports:%+v\n", ports)
	}
	err = g.startNewContainer(ctx, cli, skNet, image, &ports, env)
	if err != nil {
		ec <- err
		return
	}
	go g.deleteNoneImage(ctx, cli, image)
	fmt.Printf("Successful renew docker %s\n", strings.Split(image, "/")[1])
}

func checkParams(v url.Values) []error {
	var es []error
	var fields = []string{"pk", "image"}
	for _, f := range fields {
		val, ok := v[f]
		if !ok {
			es = append(es, fmt.Errorf("no "+f+" found in param"))
		} else {
			if f == "pk" && val[0] != os.Getenv("TOKEN") {
				es = append(es, fmt.Errorf("pk wrong"))
			}
			if f == "image" {
				s := strings.Split(val[0], "/")
				if len(s) != 3 {
					es = append(es, fmt.Errorf("image format wrong"))
				}
			}
		}
	}
	return es
}

func (g *GoHandler) deleteOldContainer(ctx context.Context, cli *dc.Client, cs types.Container) error {
	if cs.State != "exited" {
		duration := 10 * time.Second
		err := cli.ContainerStop(ctx, cs.ID[:10], &duration)
		if err != nil {
			g.logger.Printf("err:%+v\n", err)
			return err
		}
	}
	err := cli.ContainerRemove(ctx, cs.ID[:10], types.ContainerRemoveOptions{
		Force: true,
	})
	if err != nil {
		g.logger.Printf("err:%+v\n", err)
		return err
	}
	return nil
}

func (g *GoHandler) pullNewImage(ctx context.Context, cli *dc.Client, image string) error {
	rc, err := cli.ImagePull(ctx, image, types.ImagePullOptions{
		RegistryAuth: os.Getenv("SECRECT"),
	})
	if err != nil {
		g.logger.Printf("err:%+v\n", err)
		return err
	}
	if rc != nil {
		io.Copy(os.Stdout, rc)
	}
	defer rc.Close()
	return nil
}

func (g *GoHandler) getSkNetwork(ctx context.Context, cli *dc.Client) (net *types.NetworkResource, e error) {
	skNetFilter := filters.NewArgs(filters.KeyValuePair{
		Key:   "name",
		Value: "pk",
	})
	nets, err := cli.NetworkList(ctx, types.NetworkListOptions{Filters: skNetFilter})
	if err != nil {
		g.logger.Printf("err:%+v\n", err)
		return nil, err
	}
	if len(nets) == 0 {
		return nil, fmt.Errorf("no pk network found")
	}
	return &nets[0], nil
}

func (g *GoHandler) startNewContainer(ctx context.Context, cli *dc.Client, skNet *types.NetworkResource, image string, ports *nat.PortSet, env []string) error {
	if ports == nil {
		temp := make(nat.PortSet, 1)
		ports = &temp
		(*ports)["8888/tcp"] = struct{}{}
	}
	var networkConfig = nt.NetworkingConfig{
		EndpointsConfig: map[string]*nt.EndpointSettings{
			"pk": &nt.EndpointSettings{
				NetworkID: skNet.ID,
			},
		},
	}
	resp, err := cli.ContainerCreate(ctx, &container.Config{
		Image:        image,
		ExposedPorts: *ports,
		Env:          env,
	}, nil, &networkConfig, strings.Split(image, "/")[1])
	if err != nil {
		g.logger.Printf("err:%+v\n", err)
		return err
	}
	if err := cli.ContainerStart(ctx, resp.ID, types.ContainerStartOptions{}); err != nil {
		g.logger.Printf("err:%+v\n", err)
		return err
	}
	return nil
}

func (g *GoHandler) deleteNoneImage(ctx context.Context, cli *dc.Client, image string) {
	defer func() {
		if r := recover(); r != nil {
			g.logger.Println("Recovered err:", r)
		}
	}()
	imFilter := filters.NewArgs(filters.KeyValuePair{
		Key:   "dangling",
		Value: "true",
	})
	ims, err := cli.ImageList(ctx, types.ImageListOptions{
		Filters: imFilter,
	})
	if err != nil {
		g.logger.Printf("err:%+v\n", err)
	}
	for _, v := range ims {
		_, err := cli.ImageRemove(ctx, v.ID[:10], types.ImageRemoveOptions{
			Force: true,
		})
		if err != nil {
			g.logger.Printf("err:%+v\n", err)
		}
	}
}
