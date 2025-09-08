import { InfoCircleOutlined } from '@ant-design/icons';
import { Input, Tooltip } from 'antd';
import React from 'react';

import type { Notifier } from '../../../../../entity/notifiers';

interface Props {
    notifier: Notifier;
    setNotifier: (notifier: Notifier) => void;
    setIsUnsaved: (isUnsaved: boolean) => void;
}

export function EditTeamsNotifierComponent({ notifier, setNotifier, setIsUnsaved }: Props) {
    const value = notifier?.teamsNotifier?.powerAutomateUrl || '';

    const onChange = (e: React.ChangeEvent<HTMLInputElement>) => {
        const powerAutomateUrl = e.target.value.trim();
        setNotifier({
            ...notifier,
            teamsNotifier: {
                ...(notifier.teamsNotifier ?? {}),
                powerAutomateUrl,
            },
        });
        setIsUnsaved(true);
    };

    return (
        <>
            <div className="flex items-center">
                <div className="w-[130px] min-w-[130px]">Power Automate URL</div>

                <div className="w-[250px]">
                    <Input
                        value={value}
                        onChange={onChange}
                        size="small"
                        className="w-full"
                        placeholder="https://prod-00.westeurope.logic.azure.com:443/workflows/....."
                    />
                </div>

                <Tooltip
                    className="cursor-pointer"
                    title="HTTP endpoint from your Power Automate flow (When an HTTP request is received)"
                >
                    <InfoCircleOutlined className="ml-2" style={{ color: 'gray' }} />
                </Tooltip>
            </div>

            <div className="mb-1 ml-[130px] max-w-[420px] text-xs text-gray-500">
                1) In Power Automate create Flow with triggers <i>When an HTTP request is received</i>.<br />
                2) Press <i>Save</i> — you can see URL. Copy urk to this row.
            </div>
        </>
    );
}
