import { Input } from 'antd'
import { SafetyCertificateOutlined } from '@ant-design/icons'
import type { InputProps } from 'antd'

interface Props extends Omit<InputProps, 'prefix' | 'size'> {
  /** Current captcha image URL */
  captchaImg: string
  /** Click handler to refresh captcha */
  onRefresh: () => void
}

export default function CaptchaInput({ captchaImg, onRefresh, ...inputProps }: Props) {
  return (
    <div className="captcha-input">
      <Input
        prefix={<SafetyCertificateOutlined />}
        placeholder="Enter captcha code"
        maxLength={6}
        size="large"
        {...inputProps}
        style={{
          height: 52,
          fontSize: 20,
          fontWeight: 600,
          letterSpacing: '0.15em',
          textAlign: 'center',
          borderRadius: 10,
          borderWidth: 2,
          fontFamily: "'Courier New', Courier, monospace",
          ...inputProps.style,
        }}
      />
      <div style={{ marginTop: 10 }}>
        <img
          src={captchaImg}
          alt="captcha"
          onClick={onRefresh}
          style={{
            cursor: 'pointer',
            height: 52,
            width: '100%',
            objectFit: 'cover',
            borderRadius: 10,
            border: '2px solid #d9d9d9',
            transition: 'border-color 0.2s',
          }}
          title="Click to refresh captcha"
        />
        <span
          style={{
            display: 'block',
            textAlign: 'center',
            fontSize: 12,
            color: '#999',
            marginTop: 4,
          }}
        >
          Click image to refresh
        </span>
      </div>
    </div>
  )
}
