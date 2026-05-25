import { useState } from 'react';
import { useNavigate } from 'react-router-dom';
import { useAuth } from '../contexts/AuthContext';
import StatusMessage from '../components/StatusMessage';

export default function LoginPage() {
  const [isRegister, setIsRegister] = useState(false);
  const [username, setUsername] = useState('');
  const [password, setPassword] = useState('');
  const [error, setError] = useState('');
  const [loading, setLoading] = useState(false);
  const { login, register } = useAuth();
  const navigate = useNavigate();

  async function handleSubmit() {
    const u = username.trim();
    if (!u || !password) {
      setError('请填写用户名和密码');
      return;
    }

    setLoading(true);
    setError('');

    if (isRegister) {
      const err = await register(u, password);
      if (err) {
        setError(err);
        setLoading(false);
        return;
      }
      // 注册成功后自动登录
      const loginErr = await login(u, password);
      if (loginErr) {
        setError(loginErr);
      } else {
        navigate('/', { replace: true });
      }
    } else {
      const err = await login(u, password);
      if (err) {
        setError(err);
      } else {
        navigate('/', { replace: true });
      }
    }

    setLoading(false);
  }

  function handleKeyDown(e: React.KeyboardEvent) {
    if (e.key === 'Enter') handleSubmit();
  }

  return (
    <div className="page auth-page">
      <header className="header">
        <button className="back-btn" onClick={() => navigate(-1)}>← 返回</button>
        <h1>阅读器</h1>
      </header>

      <div className="auth-form">
        <div className="auth-tabs">
          <button
            className={`auth-tab ${!isRegister ? 'active' : ''}`}
            onClick={() => { setIsRegister(false); setError(''); }}
          >
            登录
          </button>
          <button
            className={`auth-tab ${isRegister ? 'active' : ''}`}
            onClick={() => { setIsRegister(true); setError(''); }}
          >
            注册
          </button>
        </div>

        <input
          type="text"
          placeholder="用户名"
          value={username}
          onChange={(e) => setUsername(e.target.value)}
          onKeyDown={handleKeyDown}
          autoFocus
        />
        <input
          type="password"
          placeholder="密码"
          value={password}
          onChange={(e) => setPassword(e.target.value)}
          onKeyDown={handleKeyDown}
        />

        <StatusMessage error={error} />

        <button className="auth-submit" onClick={handleSubmit} disabled={loading}>
          {loading ? '请稍候...' : isRegister ? '注册' : '登录'}
        </button>
      </div>
    </div>
  );
}
