/*
Tencent is pleased to support the open source community by making Blueking Container Service available.
Copyright (C) 2019 THL A29 Limited, a Tencent company. All rights reserved.
Licensed under the MIT License (the "License"); you may not use this file except
in compliance with the License. You may obtain a copy of the License at
http://opensource.org/licenses/MIT
Unless required by applicable law or agreed to in writing, software distributed under
the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
either express or implied. See the License for the specific language governing permissions and
limitations under the License.
*/

package commit

import (
	"context"
	"errors"
	"fmt"

	"github.com/spf13/viper"

	"bk-bscp/internal/database"
	pb "bk-bscp/internal/protocol/accessserver"
	pbbusinessserver "bk-bscp/internal/protocol/businessserver"
	pbcommon "bk-bscp/internal/protocol/common"
	"bk-bscp/pkg/logger"
)

// UpdateAction updates target commit object.
type UpdateAction struct {
	viper    *viper.Viper
	buSvrCli pbbusinessserver.BusinessClient

	req  *pb.UpdateCommitReq
	resp *pb.UpdateCommitResp
}

// NewUpdateAction creates new UpdateAction.
func NewUpdateAction(viper *viper.Viper, buSvrCli pbbusinessserver.BusinessClient,
	req *pb.UpdateCommitReq, resp *pb.UpdateCommitResp) *UpdateAction {
	action := &UpdateAction{viper: viper, buSvrCli: buSvrCli, req: req, resp: resp}

	action.resp.Seq = req.Seq
	action.resp.ErrCode = pbcommon.ErrCode_E_OK
	action.resp.ErrMsg = "OK"

	return action
}

// Err setup error code message in response and return the error.
func (act *UpdateAction) Err(errCode pbcommon.ErrCode, errMsg string) error {
	act.resp.ErrCode = errCode
	act.resp.ErrMsg = errMsg
	return errors.New(errMsg)
}

// Input handles the input messages.
func (act *UpdateAction) Input() error {
	if err := act.verify(); err != nil {
		return act.Err(pbcommon.ErrCode_E_AS_PARAMS_INVALID, err.Error())
	}
	return nil
}

// Output handles the output messages.
func (act *UpdateAction) Output() error {
	// do nothing.
	return nil
}

func (act *UpdateAction) verify() error {
	length := len(act.req.Bid)
	if length == 0 {
		return errors.New("invalid params, bid missing")
	}
	if length > database.BSCPIDLENLIMIT {
		return errors.New("invalid params, bid too long")
	}

	length = len(act.req.Commitid)
	if length == 0 {
		return errors.New("invalid params, commitid missing")
	}
	if length > database.BSCPIDLENLIMIT {
		return errors.New("invalid params, commitid too long")
	}

	length = len(act.req.Operator)
	if length == 0 {
		return errors.New("invalid params, operator missing")
	}
	if length > database.BSCPNAMELENLIMIT {
		return errors.New("invalid params, operator too long")
	}

	if act.req.Configs == nil {
		act.req.Configs = []byte{}
	}

	if len(act.req.Configs) > database.BSCPCONFIGSSIZELIMIT {
		return errors.New("invalid params, configs content too big")
	}
	if len(act.req.Changes) > database.BSCPCHANGESSIZELIMIT {
		return errors.New("invalid params, configs changes too big")
	}

	if len(act.req.Templateid) > database.BSCPIDLENLIMIT {
		return errors.New("invalid params, templateid too long")
	}
	if len(act.req.Template) > database.BSCPTPLSIZELIMIT {
		return errors.New("invalid params, template size too big")
	}
	if len(act.req.TemplateRule) > database.BSCPTPLRULESSIZELIMIT {
		return errors.New("invalid params, template rules too long")
	}

	if len(act.req.Configs) == 0 && len(act.req.Template) == 0 {
		return errors.New("invalid params, empty configs and template")
	}
	if len(act.req.Configs) != 0 && len(act.req.Template) != 0 {
		return errors.New("invalid params, configs and template concurrence")
	}
	if len(act.req.Template) != 0 && len(act.req.TemplateRule) == 0 {
		return errors.New("invalid params, empty template rules")
	}

	if len(act.req.Memo) > database.BSCPLONGSTRLENLIMIT {
		return errors.New("invalid params, memo too long")
	}
	return nil
}

func (act *UpdateAction) update() (pbcommon.ErrCode, string) {
	r := &pbbusinessserver.UpdateCommitReq{
		Seq:          act.req.Seq,
		Bid:          act.req.Bid,
		Commitid:     act.req.Commitid,
		Templateid:   act.req.Templateid,
		Template:     act.req.Template,
		TemplateRule: act.req.TemplateRule,
		Configs:      act.req.Configs,
		Changes:      act.req.Changes,
		Memo:         act.req.Memo,
		Operator:     act.req.Operator,
	}

	ctx, cancel := context.WithTimeout(context.Background(), act.viper.GetDuration("businessserver.calltimeout"))
	defer cancel()

	logger.V(2).Infof("UpdateCommit[%d]| request to businessserver UpdateCommit, %+v", act.req.Seq, r)

	resp, err := act.buSvrCli.UpdateCommit(ctx, r)
	if err != nil {
		return pbcommon.ErrCode_E_AS_SYSTEM_UNKONW, fmt.Sprintf("request to businessserver UpdateCommit, %+v", err)
	}
	return resp.ErrCode, resp.ErrMsg
}

// Do makes the workflows of this action base on input messages.
func (act *UpdateAction) Do() error {
	// update commit.
	if errCode, errMsg := act.update(); errCode != pbcommon.ErrCode_E_OK {
		return act.Err(errCode, errMsg)
	}
	return nil
}
