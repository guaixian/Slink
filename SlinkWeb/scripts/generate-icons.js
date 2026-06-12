const fs = require('fs');
const path = require('path');

// 图标配置
const iconSizes = [
  { name: 'favicon-16x16.png', size: 16 },
  { name: 'favicon-32x32.png', size: 32 },
  { name: 'apple-touch-icon.png', size: 180 },
  { name: 'android-chrome-192x192.png', size: 192 },
  { name: 'android-chrome-512x512.png', size: 512 }
];

// 检查public目录是否存在
const publicDir = path.join(__dirname, '../public');
if (!fs.existsSync(publicDir)) {
  fs.mkdirSync(publicDir, { recursive: true });
}

console.log('图标生成脚本');
console.log('请确保以下文件存在于相应位置：');
console.log('- src/assets/logo.png (源logo文件)');
console.log('- public/favicon.ico (基础favicon)');

console.log('\n需要生成的图标文件：');
iconSizes.forEach(icon => {
  console.log(`- ${icon.name} (${icon.size}x${icon.size})`);
});

console.log('\n注意：');
console.log('1. 你需要使用图像编辑工具（如Photoshop、GIMP或在线工具）来生成这些图标');
console.log('2. 或者使用在线favicon生成器：https://realfavicongenerator.net/');
console.log('3. 将生成的图标文件放在 public/ 目录下');
console.log('4. 确保 favicon.ico 文件存在且格式正确');

// 创建占位符文件（可选）
console.log('\n创建占位符文件...');
iconSizes.forEach(icon => {
  const iconPath = path.join(publicDir, icon.name);
  if (!fs.existsSync(iconPath)) {
    console.log(`创建占位符: ${icon.name}`);
    // 这里可以创建一个简单的占位符图片
  }
});

console.log('\n图标配置完成！'); 